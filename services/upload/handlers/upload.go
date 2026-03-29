package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	minio "github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
	"upload/models"
)

// Handler holds dependencies for the upload handler.
type Handler struct {
	minioClient *minio.Client
	bucket      string
	kafkaWriter *kafka.Writer
	maxBytes    int64
}

// New creates a Handler with the given dependencies.
func New(mc *minio.Client, bucket string, kw *kafka.Writer, maxBytes int64) *Handler {
	return &Handler{
		minioClient: mc,
		bucket:      bucket,
		kafkaWriter: kw,
		maxBytes:    maxBytes,
	}
}

// Upload handles POST /upload.
func (h *Handler) Upload(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Success: false, Error: "Unauthorized"})
		return
	}

	// Enforce max size before parsing
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes)
	if err := c.Request.ParseMultipartForm(h.maxBytes); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{Success: false, Error: "File too large"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Missing file field"})
		return
	}
	defer file.Close()

	fileID := uuid.New().String()
	storageKey := fmt.Sprintf("raw/%s", fileID)
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Upload to MinIO with 3-retry exponential backoff
	if err := h.uploadWithRetry(c.Request.Context(), storageKey, file, header.Size, mimeType); err != nil {
		log.Printf("MinIO upload failed for %s: %v", fileID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Storage unavailable"})
		return
	}

	// Publish Kafka event
	event := models.FileUploadedEvent{
		FileID:     fileID,
		OwnerID:    userName,
		FileName:   header.Filename,
		FileSize:   header.Size,
		MimeType:   mimeType,
		StorageKey: storageKey,
	}
	if err := h.publishEvent(c.Request.Context(), event); err != nil {
		log.Printf("Kafka publish failed for %s: %v — rolling back MinIO object", fileID, err)
		_ = h.minioClient.RemoveObject(context.Background(), h.bucket, storageKey, minio.RemoveObjectOptions{})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Event publish failed"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"success": true, "fileId": fileID})
}

// uploadWithRetry streams the file to MinIO with up to 3 attempts (backoff: 1s, 2s, 4s).
func (h *Handler) uploadWithRetry(ctx context.Context, key string, src interface{ Read([]byte) (int, error) }, size int64, contentType string) error {
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	var lastErr error
	for i := 0; i <= len(delays); i++ {
		_, err := h.minioClient.PutObject(ctx, h.bucket, key, src, size, minio.PutObjectOptions{ContentType: contentType})
		if err == nil {
			return nil
		}
		lastErr = err
		if i < len(delays) {
			time.Sleep(delays[i])
		}
	}
	return lastErr
}

// publishEvent marshals and writes the event to Kafka.
func (h *Handler) publishEvent(ctx context.Context, event models.FileUploadedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.kafkaWriter.WriteMessages(ctx, kafka.Message{Value: payload})
}
