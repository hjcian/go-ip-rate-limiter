package ratelimiter

import (
	"sync"
	"time"
)

type RateLimitStatus struct {
	RatelimitLimitPerMinute int   `json:"ratelimit-limit-per-minute"`
	RatelimitLimitRemaining int   `json:"ratelimit-limit-remaining"`
	RatelimitLimitReset     int64 `json:"ratelimit-limit-reset"`
	RatelimitLimitUsed      int   `json:"ratelimit-limit-used"`
}

type RateLimiter struct {
	mu     *sync.RWMutex
	limit  int // limit per minute
	count  int
	first  time.Time
	window time.Duration // rate limit window
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		mu:     &sync.RWMutex{},
		limit:  limit,
		count:  0,
		first:  time.Now(),
		window: time.Minute, // default window is 1 minute
	}
}

func (r *RateLimiter) reset() {
	r.count = 0
	r.first = time.Now()
}

func (r *RateLimiter) snapshot() (limit int, remain int, reset int64, used int) {
	limit = r.limit
	used = r.count
	remain = limit - used
	reset = r.first.Unix()
	return
}

func (r *RateLimiter) increment() {
	if r.count < r.limit {
		r.count++
	}
}

func (r *RateLimiter) canReset() bool {
	if time.Since(r.first)-r.window > 0 {
		return true
	}
	return false
}

// Allow is a goroutine-safed method that check the rate limiter if canReset(),
// if exceed limit, or just add 1 then return
func (r *RateLimiter) Allow() (bool, *RateLimitStatus) {
	var isAllow bool
	var limit, remain, used int
	var reset int64

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.canReset() {
		// can be reset
		r.reset()
		r.increment() // add 1 for current used
		isAllow = true
	} else if r.count >= r.limit {
		isAllow = false
	} else {
		r.increment()
		isAllow = true
	}
	limit, remain, reset, used = r.snapshot()

	return isAllow, &RateLimitStatus{
		RatelimitLimitPerMinute: limit,
		RatelimitLimitRemaining: remain,
		RatelimitLimitReset:     reset,
		RatelimitLimitUsed:      used,
	}
}
