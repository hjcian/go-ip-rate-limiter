package main

import "github.com/gin-gonic/gin"

func pingEndpoint(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
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
