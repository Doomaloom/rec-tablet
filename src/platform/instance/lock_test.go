package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSingleInstanceAndRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	a, e := Acquire(path)
	if e != nil {
		t.Fatal(e)
	}
	if b, e := Acquire(path); e == nil {
		b.Close()
		t.Fatal("duplicate instance allowed")
	}
	if e = a.Close(); e != nil {
		t.Fatal(e)
	}
	b, e := Acquire(path)
	if e != nil {
		t.Fatal(e)
	}
	b.Close()
}
func TestLegacyQtLockRespected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	if e := os.WriteFile(path+".lock", []byte(fmt.Sprintf("%d\nRec Tablet\n", os.Getpid())), 0600); e != nil {
		t.Fatal(e)
	}
	if l, e := Acquire(path); e == nil {
		l.Close()
		t.Fatal("running legacy app ignored")
	}
}
