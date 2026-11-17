package runtime

import (
	"context"
	"fmt"
	"os"
	"rectablet/src/platform/instance"
	store "rectablet/src/workflows/swimming/repository"
	"strings"
	"time"
)

// Boundary notices intentionally contain counts and fixed messages only. A
// signed-out tablet never exposes a previous instructor or participant name.
type syncSummary struct{ pending, review, failures int }

func (s syncSummary) text() string {
	parts := []string{}
	if s.pending > 0 {
		parts = append(parts, fmt.Sprintf("%d earlier results waiting to upload", s.pending))
	}
	if s.review > 0 {
		parts = append(parts, fmt.Sprintf("%d earlier results need supervisor review", s.review))
	}
	if s.failures > 0 {
		parts = append(parts, "Some earlier uploads could not finish; saved work will retry")
	}
	return strings.Join(parts, " · ")
}
func boundaryNotice(db *store.Store, uploadErr, readErr error, readAttempted bool) string {
	id, err := db.SyncInstructorID()
	if err != nil {
		return "Saved locally; results status unavailable"
	}
	if id == "" {
		return "Import-only session; work saved on this device"
	}
	pending, review, err := db.SyncStatus()
	if err != nil {
		return "Saved locally; results status unavailable"
	}
	parts := []string{}
	if uploadErr != nil {
		parts = append(parts, "Saved locally; upload could not finish and will retry")
	} else if pending == 0 && review == 0 {
		parts = append(parts, "Results uploaded")
	} else if pending > 0 {
		parts = append(parts, fmt.Sprintf("%d results saved locally and waiting to upload", pending))
	}
	if review > 0 {
		parts = append(parts, fmt.Sprintf("%d results need supervisor review", review))
	}
	if readAttempted {
		if readErr != nil {
			parts = append(parts, "Website refresh unavailable; local work preserved")
		} else {
			parts = append(parts, "Website results refreshed")
		}
	}
	return strings.Join(parts, " · ")
}
func (m *profileManager) retryUploadsSummary(ctx context.Context, active string) syncSummary {
	m.mu.Lock()
	profiles := append([]savedProfile(nil), m.index.Profiles...)
	m.mu.Unlock()
	summary := syncSummary{}
	for _, p := range profiles {
		if ctx.Err() != nil {
			return summary
		}
		if p.Path == active {
			continue
		}
		if _, err := os.Stat(p.Path); err != nil {
			summary.failures++
			continue
		}
		// The process already holds the base database lock for its lifetime.
		var lock *instance.Lock
		if p.Path != m.base {
			var err error
			if lock, err = instance.Acquire(p.Path); err != nil {
				summary.failures++
				continue
			}
		}
		db, err := store.Open(p.Path)
		if err == nil {
			attemptCtx, stop := context.WithTimeout(ctx, time.Second)
			err = db.UploadPending(attemptCtx)
			stop()
			if err != nil {
				summary.failures++
			}
			pending, review, statusErr := db.SyncStatus()
			summary.pending += pending
			summary.review += review
			if statusErr != nil {
				summary.failures++
			}
			_ = db.Close()
		} else {
			summary.failures++
		}
		if lock != nil {
			_ = lock.Close()
		}
	}
	return summary
}
func (s *deviceSession) logoutAsync(done func(error)) {
	s.cancelLogin()
	s.loggingOut = true
	s.notice = "Finishing result upload; saved work stays on this device"
	worker := &deviceSession{store: s.store, lock: s.lock}
	s.store, s.lock = nil, nil
	go func() {
		err := worker.logout()
		s.dispatch(func() {
			s.loggingOut = false
			s.notice = worker.notice
			if !s.closed {
				done(err)
			}
		})
	}()
}
