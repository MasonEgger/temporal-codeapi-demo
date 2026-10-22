// ABOUTME: Tests for the per-client token bucket that limits how often one address can start a run.
// Uses a fake clock so refill timing is exact.
package main

import (
	"testing"
	"time"
)

func TestLimiterAllowsBurstThenDenies(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := NewLimiter(6, 3, func() time.Time { return now })

	for i := 0; i < 3; i++ {
		if !limiter.Allow("10.0.0.1") {
			t.Fatalf("request %d should be within the burst", i+1)
		}
	}
	if limiter.Allow("10.0.0.1") {
		t.Fatal("fourth request should be denied")
	}
}

func TestLimiterRefillsAtRatePerMinute(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := NewLimiter(6, 1, func() time.Time { return now })

	limiter.Allow("10.0.0.1")
	if limiter.Allow("10.0.0.1") {
		t.Fatal("bucket should be empty")
	}
	// Six per minute is one token every ten seconds.
	now = now.Add(10 * time.Second)
	if !limiter.Allow("10.0.0.1") {
		t.Fatal("one token should have refilled after ten seconds")
	}
}

func TestLimiterNeverExceedsBurst(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := NewLimiter(60, 2, func() time.Time { return now })

	now = now.Add(time.Hour)
	allowed := 0
	for i := 0; i < 5; i++ {
		if limiter.Allow("10.0.0.1") {
			allowed++
		}
	}
	if allowed != 2 {
		t.Fatalf("allowed %d after a long idle, want the burst of 2", allowed)
	}
}

func TestLimiterKeysAreIndependent(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := NewLimiter(6, 1, func() time.Time { return now })

	limiter.Allow("10.0.0.1")
	if !limiter.Allow("10.0.0.2") {
		t.Fatal("a different address has its own bucket")
	}
}
