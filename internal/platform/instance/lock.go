// Package instance serializes access by desktop app processes on Linux.
package instance

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"strings"
)

type Lock struct{ file *os.File }

func Acquire(database string) (*Lock, error) {
	// Respect a still-running legacy Qt app during the transition. A dead Qt PID
	// does not prevent using the database; its lock file is left untouched.
	if data, e := os.ReadFile(database + ".lock"); e == nil {
		line := strings.SplitN(string(data), "\n", 2)[0]
		if pid, e := strconv.Atoi(line); e == nil && pid > 0 {
			if e = unix.Kill(pid, 0); e == nil || e == unix.EPERM {
				return nil, fmt.Errorf("this database is open in the legacy app; close it first")
			}
		}
	}
	f, e := os.OpenFile(database+".go.lock", os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		f.Close()
		return nil, fmt.Errorf("this database is already open in Rec Tablet: %w", e)
	}
	return &Lock{f}, nil
}
func (l *Lock) Close() error {
	e := unix.Flock(int(l.file.Fd()), unix.LOCK_UN)
	closeErr := l.file.Close()
	if e != nil {
		return e
	}
	return closeErr
}
