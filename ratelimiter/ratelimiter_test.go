package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

func AssertAllow(t *testing.T, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("[isAllow] got = %v, want %v", got, want)
	}
}

func AssertRatelimitLimitUsed(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("[RatelimitLimitUsed] got = %v, want %v", got, want)
	}
}

func AssertRatelimitLimitRemaining(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("[RatelimitLimitRemaining] got = %v, want %v", got, want)
	}
}

func Test_rateLimiter_goroutine_safed(t *testing.T) {
	// t.Skip()
	limit := 10000
	r := NewRateLimiter(limit)
	var wg sync.WaitGroup
	wg.Add(limit)
	for i := 0; i < limit; i++ {
		go func() {
			defer wg.Done()
			r.Allow() // comsume all rate limit quota
		}()
	}
	wg.Wait()

	// expect all rate limits are already comsumed
	// next call will got a Not Allow
	got1, got2 := r.Allow()
	AssertAllow(t, got1, false)
	AssertRatelimitLimitUsed(t, got2.RatelimitLimitUsed, limit)
	AssertRatelimitLimitRemaining(t, got2.RatelimitLimitRemaining, 0)
}

func Test_rateLimiter_reset(t *testing.T) {
	limit := 2
	window := time.Nanosecond // testing window is 1 Millisecond
	r := func(limit int) *RateLimiter {
		return &RateLimiter{
			mu:     &sync.RWMutex{},
			limit:  limit,
			count:  0,
			first:  time.Now(),
			window: window,
		}
	}(limit)

	tests := []struct {
		name  string
		want1 bool
		want2 *RateLimitStatus
	}{
		{"1 - OK", true, &RateLimitStatus{limit, 1, time.Now().Unix(), 1}},
		{"2 - OK", true, &RateLimitStatus{limit, 1, time.Now().Unix(), 1}},
		{"3 - OK", true, &RateLimitStatus{limit, 1, time.Now().Unix(), 1}},
	}
	for _, tt := range tests {
		time.Sleep(2 * window) // wait for reset
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := r.Allow()
			AssertAllow(t, got1, tt.want1)
			AssertRatelimitLimitUsed(t, got2.RatelimitLimitUsed, tt.want2.RatelimitLimitUsed)
			AssertRatelimitLimitRemaining(t, got2.RatelimitLimitRemaining, tt.want2.RatelimitLimitRemaining)
		})
	}
}

func Test_rateLimiter_Allow(t *testing.T) {
	limit := 2
	r := NewRateLimiter(limit)

	tests := []struct {
		name  string
		want1 bool
		want2 *RateLimitStatus
	}{
		{"1 - OK", true, &RateLimitStatus{limit, 1, time.Now().Unix(), 1}},
		{"2 - OK", true, &RateLimitStatus{limit, 0, time.Now().Unix(), 2}},
		{"3 - Error", false, &RateLimitStatus{limit, 0, time.Now().Unix(), 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := r.Allow()
			AssertAllow(t, got1, tt.want1)
			AssertRatelimitLimitUsed(t, got2.RatelimitLimitUsed, tt.want2.RatelimitLimitUsed)
			AssertRatelimitLimitRemaining(t, got2.RatelimitLimitRemaining, tt.want2.RatelimitLimitRemaining)
		})
	}
}
