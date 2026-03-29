package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
	"upload/config"
	"upload/handlers"
)

func main() {
	cfg := config.Load()

	// MinIO client
	mc, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatalf("failed to create MinIO client: %v", err)
	}

	// Ensure bucket exists
	exists, err := mc.BucketExists(context.Background(), cfg.MinIOBucket)
	if err != nil {
		log.Fatalf("failed to check MinIO bucket: %v", err)
	}
	if !exists {
		if err := mc.MakeBucket(context.Background(), cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("failed to create MinIO bucket: %v", err)
		}
	}

	// Kafka writer
	kw := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   "file.uploaded",
	})
	defer kw.Close()

	h := handlers.New(mc, cfg.MinIOBucket, kw, cfg.MaxUploadBytes)

	r := gin.Default()
	r.POST("/upload", h.Upload)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Printf("Upload service starting on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
