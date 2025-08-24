package routes

import (
	"net/http"

	"auth/config"
	"auth/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, cfg config.Config, h *handlers.AuthHandler) {
	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.AllowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	ah := h.WithConfig(cfg)

	r.POST("/login", ah.Login)
	r.POST("/register", ah.Register)
	r.GET("/kong-validate", ah.KongValidate)
	r.POST("/logout", ah.Logout)
	r.GET("/health", ah.Health)
}
