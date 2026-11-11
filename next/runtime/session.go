package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"rectablet/src/platform/instance"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
)

// deviceSession owns the open instructor database. All methods and completions
// run on the UI thread; only the network fetch runs in a worker goroutine.
// The runtime holds the base database lock for the lifetime of the process.
type deviceSession struct {
	profiles   *profileManager
	server     string
	store      *store.Store
	lock       *instance.Lock
	cancel     context.CancelFunc
	generation uint64
	closed     bool
	fetch      func(context.Context, string, string) (transfer.ClassPackage, error)
	dispatch   func(func())
}

func (s *deviceSession) login(code string, done func(error)) {
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
		if err == nil && ctx.Err() == nil {
			err = candidate.open(pkg)
		}
		if err == nil && ctx.Err() == nil {
			// Every successful code login retries all instructors using upload-only
			// capabilities. Only this active profile receives a read capability.
			s.profiles.retryUploads(ctx, candidate.store.Path)
			uploadCtx, stop := context.WithTimeout(ctx, 5*time.Second)
			_ = candidate.store.UploadPending(uploadCtx)
			stop()
			readCtx, stop := context.WithTimeout(ctx, 5*time.Second)
			_ = candidate.store.ReadAccepted(readCtx, pkg.Credentials)
			stop()
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
			}
			done(err)
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
		ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.store.UploadPending(ctx)
		stop()
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
