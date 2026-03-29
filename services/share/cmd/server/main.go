package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gokusan/share/config"
	"github.com/gokusan/share/handlers"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Initialize HTTP client for internal service calls
	httpClient := &http.Client{}

	h := handlers.New(rdb, httpClient, cfg.MetadataServiceURL, cfg.DownloadServiceURL)

	r := gin.Default()

	r.POST("/share", h.CreateShare)
	r.GET("/share/:token", h.ResolveShare)
	r.DELETE("/share/:token", h.DeleteShare)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Printf("share service starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
