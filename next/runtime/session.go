package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rectablet/src/platform/instance"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
)

// deviceSession owns the open instructor database. All methods and completions
// run on the UI thread; login and logout storage/network work runs in workers.
// The runtime holds the base database lock for the lifetime of the process.
type deviceSession struct {
	profiles   *profileManager
	server     string
	store      *store.Store
	lock       *instance.Lock
	cancel     context.CancelFunc
	generation uint64
	closed     bool
	loggingOut bool
	notice     string
	fetch      func(context.Context, string, string) (transfer.ClassPackage, error)
	dispatch   func(func())
}

// savedDataRecoveredMessage tells the instructor that their earlier saved work
// on this tablet could not be found and a fresh copy of their classes was loaded.
const savedDataRecoveredMessage = "Earlier saved results on this tablet were not found. Your classes were loaded again."

// databaseOpenMessage is shown when the instructor's database is held by another
// window of the app.
const databaseOpenMessage = "Rec Tablet is already open for this instructor. Close the other window, then try again."

// plainLoginError replaces developer text from a failed login with an
// instruction, keeping the cause for errors.Is. Messages already written for
// the instructor pass through.
func plainLoginError(err error) error {
	var shared *transfer.ShareError
	if err == nil || errors.As(err, &shared) {
		return err
	}
	var network net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &network) && network.Timeout()):
		return &transfer.ShareError{Message: transfer.NoConnectionMessage, Cause: err}
	case strings.Contains(err.Error(), "already open in Rec Tablet"):
		return &transfer.ShareError{Message: databaseOpenMessage, Cause: err}
	}
	return err
}

func (s *deviceSession) login(code string, done func(error)) {
	if s.loggingOut {
		done(fmt.Errorf("finishing logout; please try the code again shortly"))
		return
	}
	if s.closed || s.store != nil || s.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.generation++
	generation := s.generation
	server := s.profiles.shareServer(s.server)
	go func() {
		pkg, err := s.fetch(ctx, server, code)
		candidate := &deviceSession{profiles: s.profiles}
		recovered := false
		if err == nil && ctx.Err() == nil {
			err = candidate.open(pkg)
			recovered = s.profiles.takeRecovered()
		}
		if err == nil && ctx.Err() == nil {
			// Every successful code login retries all instructors using upload-only
			// capabilities. Only this active profile receives a read capability.
			previous := s.profiles.retryUploadsSummary(ctx, candidate.store.Path)
			uploadCtx, stop := context.WithTimeout(ctx, uploadBudget(candidate.store))
			uploadErr := candidate.store.UploadPending(uploadCtx)
			stop()
			readCtx, stop := context.WithTimeout(ctx, 5*time.Second)
			readErr := candidate.store.ReadAccepted(readCtx, pkg.Credentials)
			stop()
			candidate.notice = boundaryNotice(candidate.store, uploadErr, readErr, pkg.Credentials != nil)
			if recovered {
				candidate.notice = strings.TrimSuffix(savedDataRecoveredMessage+" · "+candidate.notice, " · ")
			}
			if earlier := previous.text(); earlier != "" {
				candidate.notice += " · " + earlier
			}
		}
		if ctx.Err() != nil {
			candidate.release()
			return
		}
		s.dispatch(func() {
			if s.closed || ctx.Err() != nil || generation != s.generation {
				candidate.release()
				return
			}
			s.cancel = nil
			defer cancel()
			if err == nil {
				s.store, s.lock = candidate.store, candidate.lock
				s.notice = candidate.notice
			}
			done(plainLoginError(err))
		})
	}()
}

func (s *deviceSession) open(pkg transfer.ClassPackage) error {
	path := s.profiles.packagePath(pkg)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	var lock *instance.Lock
	if path != s.profiles.base {
		var err error
		lock, err = instance.Acquire(path)
		if err != nil {
			return err
		}
	}
	release := func() {
		if lock != nil {
			_ = lock.Close()
		}
	}
	profile, err := s.profiles.importPackage(pkg)
	if err != nil {
		release()
		return err
	}
	if profile.Path != path {
		// An incompatible legacy profile was preserved; lock the database in use.
		release()
		lock = nil
		if profile.Path != s.profiles.base {
			if lock, err = instance.Acquire(profile.Path); err != nil {
				return err
			}
		}
	}
	database, err := store.Open(profile.Path)
	if err != nil {
		release()
		return err
	}
	if err = s.profiles.activate(profile.Key); err != nil {
		_ = database.Close()
		release()
		return err
	}
	s.store, s.lock = database, lock
	return nil
}

func (s *deviceSession) cancelLogin() {
	s.generation++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *deviceSession) logout() error {
	s.cancelLogin()
	var err error
	if s.store != nil {
		ctx, stop := context.WithTimeout(context.Background(), uploadBudget(s.store))
		uploadErr := s.store.UploadPending(ctx)
		stop()
		s.notice = boundaryNotice(s.store, uploadErr, nil, false)
		err = s.store.Close()
		s.store = nil
	}
	if s.lock != nil {
		if e := s.lock.Close(); err == nil {
			err = e
		}
		s.lock = nil
	}
	if err != nil {
		return fmt.Errorf("could not close the instructor session: %w", err)
	}
	return nil
}
func (s *deviceSession) close() { s.closed = true; _ = s.logout() }

// release discards an unmounted or cancelled login without retaining a read token.
func (s *deviceSession) release() {
	if s.store != nil {
		_ = s.store.Close()
		s.store = nil
	}
	if s.lock != nil {
		_ = s.lock.Close()
		s.lock = nil
	}
}

// Local databases remain durable but cannot authorize a new session.
func (s *deviceSession) resume() error { return fmt.Errorf("online instructor code required") }
