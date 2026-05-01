package retry

import (
	"container/heap"
	"time"
)

type Item[T any] struct {
	Value     T
	ReadyAt   time.Time
	Attempts  int
	lastIndex int
}

type Queue[T any] []*Item[T]

func NewQueue[T any]() *Queue[T] {
	q := Queue[T]{}
	heap.Init(&q)
	return &q
}

func (q Queue[T]) Len() int { return len(q) }

func (q Queue[T]) Less(i, j int) bool {
	return q[i].ReadyAt.Before(q[j].ReadyAt)
}

func (q Queue[T]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].lastIndex = i
	q[j].lastIndex = j
}

func (q *Queue[T]) Push(x any) {
	item := x.(*Item[T])
	item.lastIndex = len(*q)
	*q = append(*q, item)
}

func (q *Queue[T]) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*q = old[:n-1]
	return item
}

func (q *Queue[T]) Schedule(value T, attempts int, delay time.Duration) {
	heap.Push(q, &Item[T]{Value: value, Attempts: attempts, ReadyAt: time.Now().Add(delay)})
}

func (q *Queue[T]) PopReady(now time.Time) (T, int, bool) {
	var zero T
	if q.Len() == 0 || (*q)[0].ReadyAt.After(now) {
		return zero, 0, false
	}
	item := heap.Pop(q).(*Item[T])
	return item.Value, item.Attempts, true
}
