package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		if ctx.Err() != nil {
			return
		}
		s.dispatch(func() {
			if s.closed || ctx.Err() != nil || generation != s.generation {
				return
			}
			s.cancel = nil
			defer cancel()
			if err == nil {
				err = s.open(pkg)
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

// resume opens only the last activated local profile without redeeming a code.
func (s *deviceSession) resume() error {
	if s.closed || s.store != nil || s.cancel != nil {
		return fmt.Errorf("cannot restore a session while another login is active")
	}
	profile, ok := s.profiles.lastProfile()
	if !ok {
		return fmt.Errorf("no previous session is available; log in with a code first")
	}
	if _, err := os.Stat(profile.Path); err != nil {
		return fmt.Errorf("previous session is unavailable: %w", err)
	}
	var lock *instance.Lock
	if profile.Path != s.profiles.base {
		var err error
		lock, err = instance.Acquire(profile.Path)
		if err != nil {
			return err
		}
	}
	release := func() {
		if lock != nil {
			_ = lock.Close()
		}
	}
	database, err := store.Open(profile.Path)
	if err != nil {
		release()
		return err
	}
	dataset, err := database.Meta("dataset_id")
	if err == nil {
		var instructor string
		instructor, err = database.Meta("instructor")
		if err == nil && (dataset != profile.DatasetID || instructor != profile.Instructor) {
			err = fmt.Errorf("previous session does not match the saved instructor")
		}
	}
	if err != nil {
		_ = database.Close()
		release()
		return err
	}
	s.store, s.lock = database, lock
	return nil
}
