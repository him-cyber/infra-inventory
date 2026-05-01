package ratelimit

import "testing"

func TestLimiterExhaustsTenantBucket(t *testing.T) {
	limiter := NewLimiter()
	if !limiter.Allow("demo", 1) {
		t.Fatal("first request should pass")
	}
	if limiter.Allow("demo", 1) {
		t.Fatal("second immediate request should be limited")
	}
}
