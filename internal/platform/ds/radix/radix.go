package radix

import (
	"sort"
	"strings"
)

type Radix[T any] struct {
	entries []entry[T]
}

type entry[T any] struct {
	prefix string
	value  T
}

func New[T any](values map[string]T) *Radix[T] {
	entries := make([]entry[T], 0, len(values))
	for prefix, value := range values {
		entries = append(entries, entry[T]{prefix: strings.Trim(prefix, "/"), value: value})
	}
	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].prefix) > len(entries[j].prefix)
	})
	return &Radix[T]{entries: entries}
}

func (r *Radix[T]) LongestPrefix(key string) (T, bool) {
	key = strings.Trim(key, "/")
	var zero T
	if r == nil {
		return zero, false
	}
	for _, e := range r.entries {
		if key == e.prefix || strings.HasPrefix(key, e.prefix+"/") {
			return e.value, true
		}
	}
	return zero, false
}
