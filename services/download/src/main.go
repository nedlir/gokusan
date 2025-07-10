package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/download", func(c *gin.Context) {
		c.String(http.StatusOK, "Ok from download")
	})

	r.Run(":8012")
}
