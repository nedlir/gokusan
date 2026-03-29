package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"download/config"
	"download/handlers"
)

func main() {
	cfg := config.Load()

	mc, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatalf("failed to create MinIO client: %v", err)
	}

	h := handlers.New(mc, cfg.MinIOBucket, cfg.MetadataServiceURL)

	r := gin.Default()
	r.GET("/download/:id", h.Download)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Printf("Download service starting on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
