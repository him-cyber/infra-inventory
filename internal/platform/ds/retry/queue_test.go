package retry

import (
	"testing"
	"time"
)

func TestQueuePopsEarliestReadyItem(t *testing.T) {
	q := NewQueue[string]()
	q.Schedule("later", 1, time.Hour)
	q.Schedule("now", 1, -time.Second)
	value, attempts, ok := q.PopReady(time.Now())
	if !ok {
		t.Fatal("expected ready item")
	}
	if value != "now" || attempts != 1 {
		t.Fatalf("got value=%q attempts=%d", value, attempts)
	}
}
