package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	capacity float64
	tokens   float64
	rate     float64
	seen     time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{buckets: make(map[string]*bucket)}
}

func (l *Limiter) Allow(key string, perMinute int) bool {
	if perMinute <= 0 {
		perMinute = 60
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{capacity: float64(perMinute), tokens: float64(perMinute), rate: float64(perMinute) / 60.0, seen: now}
		l.buckets[key] = b
	}
	elapsed := now.Sub(b.seen).Seconds()
	b.tokens = min(b.capacity, b.tokens+elapsed*b.rate)
	b.seen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
