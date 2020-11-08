package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitStatus struct {
	RatelimitLimitPerMinute int   `json:"ratelimit-limit-per-minute"`
	RatelimitLimitRemaining int   `json:"ratelimit-limit-remaining"`
	RatelimitLimitReset     int64 `json:"ratelimit-limit-reset"`
	RatelimitLimitUsed      int   `json:"ratelimit-limit-used"`
}

type rateLimiter struct {
	mu    *sync.RWMutex
	limit int // limit per minute
	count int
	first time.Time
}

func newRateLimiter(limit int) *rateLimiter {
	return &rateLimiter{
		mu:    &sync.RWMutex{},
		limit: limit,
		count: 0,
		first: time.Now(),
	}
}

func (r *rateLimiter) reset() {
	r.count = 0
	r.first = time.Now()
}

func (r *rateLimiter) snapshot() (limit int, remain int, reset int64, used int) {
	limit = r.limit
	used = r.count
	remain = limit - used
	reset = r.first.Unix()
	return
}

func (r *rateLimiter) increment() {
	if r.count < r.limit {
		r.count++
	}
}

func (r *rateLimiter) Allow() (bool, *rateLimitStatus) {
	var isAllow bool
	var limit, remain, used int
	var reset int64

	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Since(r.first) > time.Minute {
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

	return isAllow, &rateLimitStatus{
		RatelimitLimitPerMinute: limit,
		RatelimitLimitRemaining: remain,
		RatelimitLimitReset:     reset,
		RatelimitLimitUsed:      used,
	}
}

// IPRateLimiter is a goroutine-safed IPRateLimiter object
type IPRateLimiter struct {
	ips            map[string]*rateLimiter
	mu             *sync.RWMutex
	reqLimitPerMin int
}

// NewIPRateLimiter create a goroutine-safed IPRateLimiter object
func NewIPRateLimiter(reqLimitPerMin int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:            make(map[string]*rateLimiter),
		mu:             &sync.RWMutex{},
		reqLimitPerMin: reqLimitPerMin,
	}
}

func (i *IPRateLimiter) addIP(ip string) *rateLimiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := newRateLimiter(i.reqLimitPerMin)
	i.ips[ip] = limiter
	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rateLimiter {
	i.mu.Lock()
	limiter, isExists := i.ips[ip]
	if !isExists {
		i.mu.Unlock()
		return i.addIP(ip)
	}
	i.mu.Unlock()

	return limiter
}

const RequestLimitPerMinute = 60

func pingEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": 1996,
	})
}

func setHeaders(c *gin.Context, ip string, statusSnapshot *rateLimitStatus) {
	c.Writer.Header().Set("X-ratelimit-limit-ip", fmt.Sprint(ip))
	c.Writer.Header().Set("X-ratelimit-limit-per-minute", fmt.Sprint(statusSnapshot.RatelimitLimitPerMinute))
	c.Writer.Header().Set("X-ratelimit-limit-remaining", fmt.Sprint(statusSnapshot.RatelimitLimitRemaining))
	c.Writer.Header().Set("X-ratelimit-limit-reset", fmt.Sprint(statusSnapshot.RatelimitLimitReset))
	c.Writer.Header().Set("X-ratelimit-limit-used", fmt.Sprint(statusSnapshot.RatelimitLimitUsed))
}

// RateLimiter is a middleware to limit the request rate
func RateLimiter() gin.HandlerFunc {
	var ipLimiter = NewIPRateLimiter(RequestLimitPerMinute)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := ipLimiter.GetLimiter(ip)
		isAllow, statusSnapshot := limiter.Allow()
		setHeaders(c, ip, statusSnapshot)
		if isAllow {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}

// The engine with all endpoints is now extracted from the main function
func setupServer() *gin.Engine {
	// https://github.com/gin-gonic/gin#blank-gin-without-middleware-by-default
	server := gin.New()
	server.Use(RateLimiter())

	// NoRoute to simple match all requests
	server.NoRoute(pingEndpoint)

	return server
}

func main() {
	setupServer().Run()
}
