package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/upload", func(c *gin.Context) {
		// TODO: Currently trusting headers without validation
		// TODO: Kong is hardcoding admin headers instead of validating JWTs
		// TODO: obviously tthese headers can be spoofed if Kong auth is bypassed
		userName := c.GetHeader("X-User-Name")
		userRole := c.GetHeader("X-User-Role")

		if userName == "" || userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("Upload successful! Welcome %s (role: %s)", userName, userRole))
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Println("Upload service starting on :6565")
	r.Run(":6565")
}
