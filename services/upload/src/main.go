package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/upload", func(c *gin.Context) {
		c.String(http.StatusOK, "Ok from Upload!")
	})

	r.Run(":6565")
}
