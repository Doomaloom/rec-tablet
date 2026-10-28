package navigation

import (
	"fmt"
	"sync"

	"rectablet/next/contracts"
)

// Edit contains an explicit saved value, never a toggle evaluated by the worker.
// Preparation and painting run on the UI thread; Save must only use storage.
type Edit struct {
	Row    Row
	Button contracts.ButtonUpdate
	Save   func() error
}

type editor interface{ PrepareEdit(string) (Edit, bool) }

func (c *Controller) SubscribeEdits(fn func(contracts.ButtonUpdate)) { c.onEdit = fn }

// EnableAsyncSaves uses one worker to preserve tap order. dispatch schedules UI work.
func (c *Controller) EnableAsyncSaves(dispatch func(func())) {
	if c.writes != nil {
		return
	}
	c.writes = newWriteQueue(func(err error) { dispatch(func() { c.finishWrite(err) }) })
}

func (c *Controller) quickEdit(id string) (bool, error) {
	if c.Closed {
		return true, nil
	}
	if len(c.afterWrites) > 0 {
		return true, nil
	}
	if c.onEdit == nil && c.writes == nil {
		return false, nil
	}
	feature, err := c.active()
	if err != nil {
		return false, err
	}
	e, ok := feature.(editor)
	if !ok {
		return false, nil
	}
	edit, ok := e.PrepareEdit(id)
	if !ok {
		return false, nil
	}
	// A nil Save means the feature intentionally keeps edits local until Save student.
	if edit.Save != nil {
		if c.writes == nil {
			if err := edit.Save(); err != nil {
				return true, c.Fail(fmt.Errorf("could not save change: %w", err))
			}
		} else {
			c.pendingWrites++
			c.writes.add(edit.Save)
		}
	}
	hadNotice := c.State.Notice != ""
	c.State.Notice, c.State.Error = "", false
	for i, row := range c.State.Rows {
		if row.ID == edit.Row.ID {
			c.State.Rows[i] = edit.Row
			break
		}
	}
	if c.onEdit != nil {
		c.onEdit(edit.Button)
	} else {
		c.changed()
	}
	if hadNotice && c.onEdit != nil {
		c.changed()
	}
	return true, nil
}

func (c *Controller) finishWrite(err error) {
	if c.Closed {
		return
	}
	c.pendingWrites--
	if err != nil && c.saveError == nil {
		c.saveError = err
	}
	if c.pendingWrites != 0 {
		return
	}
	actions := c.afterWrites
	c.afterWrites = nil
	if c.saveError != nil {
		err := c.saveError
		c.saveError = nil
		// Reload persisted values and stay on the current page so failed changes
		// are never silently submitted or discarded by deferred navigation.
		_ = c.Refresh()
		c.Fail(fmt.Errorf("could not save change; check your marks and try again: %w", err))
		return
	}
	for _, action := range actions {
		if err := c.Run(action); err != nil {
			break
		}
	}
}

// DrainSaves is used before closing the store. It does not wait on UI callbacks.
func (c *Controller) DrainSaves() {
	if c.writes != nil {
		c.writes.close()
	}
}

type writeQueue struct {
	mu       sync.Mutex
	jobs     []func() error
	wake     chan struct{}
	done     chan struct{}
	closed   bool
	complete func(error)
}

func newWriteQueue(complete func(error)) *writeQueue {
	q := &writeQueue{wake: make(chan struct{}, 1), done: make(chan struct{}), complete: complete}
	go q.run()
	return q
}
func (q *writeQueue) add(job func() error) {
	q.mu.Lock()
	q.jobs = append(q.jobs, job)
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
}
func (q *writeQueue) run() {
	defer close(q.done)
	for {
		q.mu.Lock()
		if len(q.jobs) == 0 {
			closed := q.closed
			q.mu.Unlock()
			if closed {
				return
			}
			<-q.wake
			continue
		}
		job := q.jobs[0]
		q.jobs[0] = nil
		q.jobs = q.jobs[1:]
		q.mu.Unlock()
		q.complete(job())
	}
}
func (q *writeQueue) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
	<-q.done
}
