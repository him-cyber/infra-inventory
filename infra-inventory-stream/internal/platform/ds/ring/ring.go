package ring

import "sync"

type Buffer[T any] struct {
	mu     sync.RWMutex
	values []T
	next   int
	full   bool
}

func New[T any](size int) *Buffer[T] {
	if size <= 0 {
		size = 1
	}
	return &Buffer[T]{values: make([]T, size)}
}

func (b *Buffer[T]) Push(value T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.values[b.next] = value
	b.next = (b.next + 1) % len(b.values)
	if b.next == 0 {
		b.full = true
	}
}

func (b *Buffer[T]) Snapshot() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.full {
		out := make([]T, b.next)
		copy(out, b.values[:b.next])
		return out
	}
	out := make([]T, 0, len(b.values))
	out = append(out, b.values[b.next:]...)
	out = append(out, b.values[:b.next]...)
	return out
}
