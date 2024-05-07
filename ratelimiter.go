// Sliding Window algorithm chosen for its efficiency with bursts while maintaining accuracy.
// Fixed Window might not adapt well to bursts, while Leaky Bucket and Token Bucket might introduce delays
// even during non-burst periods.

package main

import (
	"sync"
	"time"
)

type RateLimiter interface {
	ResetWindow()
	AllowRequest() time.Time
}

type RateLimiterConfig struct {
	Rate   int // Maximum requests per second
	Window time.Duration
}

func NewRateLimiterConfig(rate int) *RateLimiterConfig {
	return &RateLimiterConfig{
		Rate:   rate,            // No of requests allowed per second
		Window: time.Second * 1, // Interval to reset window
	}
}

// WindowTracker tracks requests within the window
type WindowTracker struct {
	mu           sync.Mutex
	StartTime    time.Time
	RequestCount int
	config       *RateLimiterConfig
}

func NewWindowTracker(rlconfig *RateLimiterConfig) *WindowTracker {
	return &WindowTracker{
		mu:           sync.Mutex{},
		StartTime:    time.Now(),
		RequestCount: 0,
		config:       rlconfig,
	}
}

func (wt *WindowTracker) ResetWindow() {
	wt.StartTime = time.Now()
	wt.RequestCount = 0
}

func (wt *WindowTracker) AllowRequest() bool {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if time.Now().After(wt.StartTime.Add(wt.config.Window)) {
		wt.ResetWindow()
	}

	if wt.RequestCount < wt.config.Rate {
		wt.RequestCount++
		return true
	}
	return false
}
