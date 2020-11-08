package ipratelimiter

import (
	"goipratelimiter/ratelimiter"
	"sync"
)

// IPRateLimiter is a goroutine-safed IPRateLimiter object
type IPRateLimiter struct {
	ips            map[string]*ratelimiter.RateLimiter
	mu             *sync.RWMutex
	reqLimitPerMin int
}

// NewIPRateLimiter create a goroutine-safed IPRateLimiter object
func NewIPRateLimiter(reqLimitPerMin int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:            make(map[string]*ratelimiter.RateLimiter),
		mu:             &sync.RWMutex{},
		reqLimitPerMin: reqLimitPerMin,
	}
}

func (i *IPRateLimiter) totalIP() int {
	return len(i.ips)
}

func (i *IPRateLimiter) addIP(ip string) *ratelimiter.RateLimiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := ratelimiter.NewRateLimiter(i.reqLimitPerMin)
	i.ips[ip] = limiter
	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *ratelimiter.RateLimiter {
	i.mu.Lock()
	limiter, isExists := i.ips[ip]
	if !isExists {
		i.mu.Unlock()
		return i.addIP(ip)
	}
	i.mu.Unlock()

	return limiter
}
