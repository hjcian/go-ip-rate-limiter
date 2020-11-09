package main

import (
	"fmt"
	"goipratelimiter/ipratelimiter"
	"goipratelimiter/ratelimiter"
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestLimitPerMinute = 60

func pingEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": 1996,
	})
}

func setHeaders(c *gin.Context, ip string, statusSnapshot *ratelimiter.RateLimitStatus) {
	c.Writer.Header().Set("X-ratelimit-limit-ip", fmt.Sprint(ip))
	c.Writer.Header().Set("X-ratelimit-limit-per-minute", fmt.Sprint(statusSnapshot.RatelimitLimitPerMinute))
	c.Writer.Header().Set("X-ratelimit-limit-remaining", fmt.Sprint(statusSnapshot.RatelimitLimitRemaining))
	c.Writer.Header().Set("X-ratelimit-limit-reset", fmt.Sprint(statusSnapshot.RatelimitLimitReset))
	c.Writer.Header().Set("X-ratelimit-limit-used", fmt.Sprint(statusSnapshot.RatelimitLimitUsed))
}

// RateLimiter is a middleware to limit the request rate
func RateLimiter(requestLimitPerMin int) gin.HandlerFunc {
	var ipLimiter = ipratelimiter.NewIPRateLimiter(requestLimitPerMin)
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
func setupServer(requestLimitPerMin int) *gin.Engine {
	// https://github.com/gin-gonic/gin#blank-gin-without-middleware-by-default
	server := gin.New()
	if requestLimitPerMin >= 0 {
		// if requestLimitPerMin < 0, means NO rate limit
		server.Use(RateLimiter(requestLimitPerMin))
	}

	// NoRoute to simple match all requests
	server.NoRoute(pingEndpoint)

	return server
}

func main() {
	setupServer(RequestLimitPerMinute).Run()
}
