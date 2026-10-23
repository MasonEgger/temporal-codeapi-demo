// ABOUTME: Per-key token bucket used to cap how many runs one client address can start per minute.
// Buckets live in memory, like sessions; a gateway restart forgets them.
package main

import (
	"sync"
	"time"
)

// Limiter hands out tokens per key at a steady rate up to a burst.
type Limiter struct {
	mu      sync.Mutex
	perMin  float64
	burst   float64
	now     func() time.Time
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewLimiter allows perMinute requests per key on average and at most burst at once.
// now is injectable for tests; pass time.Now otherwise.
func NewLimiter(perMinute, burst int, now func() time.Time) *Limiter {
	return &Limiter{perMin: float64(perMinute), burst: float64(burst), now: now, buckets: map[string]*bucket{}}
}

// Allow reports whether key may proceed, spending one token if so.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	current := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: current}
		l.buckets[key] = b
	}
	b.tokens += current.Sub(b.last).Minutes() * l.perMin
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = current
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
