package runtime

import (
	"fmt"
	"testing"
	"time"
)

func TestUploadBudgetScalesWithQueueAndIsCapped(t *testing.T) {
	s, _ := sessionFixture(t)
	checkSession(t, s.open(boundProfile("Alex", "a", "http://127.0.0.1:1")))
	defer s.release()
	if got := uploadBudget(s.store); got != 5*time.Second {
		t.Fatalf("empty queue budget %v", got)
	}
	queueRows := func(from, to int) {
		for i := from; i < to; i++ {
			_, err := s.store.DB.Exec("INSERT INTO sync_outbox(operation_id,record_id,body) VALUES(?,?,'{}')", fmt.Sprintf("op-%d", i), fmt.Sprintf("record-%d", i))
			checkSession(t, err)
		}
	}
	queueRows(0, 30)
	if got := uploadBudget(s.store); got != 8*time.Second {
		t.Fatalf("30 pending budget %v", got)
	}
	queueRows(30, 1000)
	if got := uploadBudget(s.store); got != 20*time.Second {
		t.Fatalf("budget not capped: %v", got)
	}
}

func TestEarlierUploadBudgetStartsAtOneSecondAndScalesWithQueue(t *testing.T) {
	s, _ := sessionFixture(t)
	checkSession(t, s.open(boundProfile("Alex", "a", "http://127.0.0.1:1")))
	defer s.release()
	if got := earlierUploadBudget(s.store); got != time.Second {
		t.Fatalf("empty queue budget %v", got)
	}
	for i := 0; i < 1000; i++ {
		_, err := s.store.DB.Exec("INSERT INTO sync_outbox(operation_id,record_id,body) VALUES(?,?,'{}')", fmt.Sprintf("op-%d", i), fmt.Sprintf("record-%d", i))
		checkSession(t, err)
		if i == 29 {
			if got := earlierUploadBudget(s.store); got != 4*time.Second {
				t.Fatalf("30 pending budget %v", got)
			}
		}
	}
	if got := earlierUploadBudget(s.store); got != 10*time.Second {
		t.Fatalf("budget not capped: %v", got)
	}
}
