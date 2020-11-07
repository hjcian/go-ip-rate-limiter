package main

import "github.com/gin-gonic/gin"

// RateLimitPerMinute as its name
const RateLimitPerMinute = 60

func pingEndpoint(c *gin.Context) {
	c.JSON(200, gin.H{
		"ip":                         c.ClientIP(), // leverage the behavior of c.ClientIP()
		"ratelimit-limit-per-minute": RateLimitPerMinute,
		"ratelimit-limit-remaining":  RateLimitPerMinute,
		"ratelimit-limit-reset":      "123456789",
		"ratelimit-limit-used":       0,
	})
}

// The engine with all endpoints is now extracted from the main function
func setupServer() *gin.Engine {
	server := gin.Default()

	// NoRoute to simple match all requests
	server.NoRoute(pingEndpoint)

	return server
}

func main() {
	setupServer().Run()
}
