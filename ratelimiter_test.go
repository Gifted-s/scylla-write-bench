package main

import (
	"sync"
	"testing"
	"time"
)

// Normally we would use a library that supports time mocking like testify (https://github.com/stretchr/testify)
// but requirement does not allow external libraries

func TestResetWindow(t *testing.T) {
	test_rate := 40
	test_window := 1 * time.Second

	mockTime := time.Now() // Replace with mocking library

	wt := &WindowTracker{
		mu:           sync.Mutex{},
		StartTime:    mockTime,
		RequestCount: 0,
		config: &RateLimiterConfig{
			Rate:   test_rate,
			Window: test_window,
		},
	}
	// Fill window
	wt.RequestCount = 0
	for i := 0; i < test_rate; i++ {
		wt.AllowRequest()
	}

	// ensure window is full
	allowed := wt.AllowRequest()
	if allowed || wt.RequestCount != test_rate {
		t.Errorf("Expected request denied after reaching rate, got allowed: %v, request count: %v", allowed, wt.RequestCount)
	}
	// reset window
	wt.ResetWindow()

	if wt.RequestCount != 0 {
		t.Errorf("Expected request count to be reset to zero but found:request count: %v", wt.RequestCount)
	}

	now := time.Now()
	if now.Sub(wt.StartTime) < 0 {
		t.Errorf("Expected start time to be reset to now but found: %v", wt.StartTime)
	}

}

func TestAllowRequest(t *testing.T) {
	test_rate := 40
	test_window := 1 * time.Second

	mockTime := time.Now() // Replace with mocking library

	wt := &WindowTracker{
		mu:           sync.Mutex{},
		StartTime:    mockTime,
		RequestCount: 0,
		config: &RateLimiterConfig{
			Rate:   test_rate,
			Window: test_window,
		},
	}

	// Test within Window
	allowed := wt.AllowRequest()
	if !allowed || wt.RequestCount != 1 {
		t.Errorf("Expected request allowed within window, got allowed: %v, request count: %v", allowed, wt.RequestCount)
	}

	// Fill window
	wt.RequestCount = 0
	for i := 0; i < test_rate; i++ {
		wt.AllowRequest()
	}
	allowed = wt.AllowRequest()
	if allowed || wt.RequestCount != test_rate {
		t.Errorf("Expected request denied after reaching rate, got allowed: %v, request count: %v", allowed, wt.RequestCount)
	}

	// Sleep to allow refill
	time.Sleep(time.Second * 2)

	// Test after Window is reset
	allowed = wt.AllowRequest()
	if !allowed || wt.RequestCount != 1 {
		t.Errorf("Expected request allowed after window reset, got allowed: %v, request count: %v", allowed, wt.RequestCount)
	}
}
