package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gokusan/metadata/models"
	"github.com/gokusan/metadata/repository"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	repo        *repository.Repository
	kafkaWriter *kafka.Writer
}

// New creates a Handler and initialises the Kafka writer for the file.deleted topic.
func New(repo *repository.Repository, broker string) *Handler {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   "file.deleted",
	})
	return &Handler{repo: repo, kafkaWriter: w}
}

// Health returns a simple liveness response.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// ListFiles handles GET /files — returns all non-deleted files for the requesting user.
func (h *Handler) ListFiles(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "missing X-User-Name header"})
		return
	}

	files, err := h.repo.GetFilesByOwner(c.Request.Context(), userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "files": files})
}

// GetFile handles GET /files/:id — returns file metadata; 403 if not owner.
func (h *Handler) GetFile(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "missing X-User-Name header"})
		return
	}

	id := c.Param("id")
	file, err := h.repo.GetFileByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	if file.OwnerID != userName {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Success: false, Error: "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "file": file})
}

// DeleteFile handles DELETE /files/:id — soft-deletes the file and publishes a file.deleted event.
func (h *Handler) DeleteFile(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "missing X-User-Name header"})
		return
	}

	id := c.Param("id")
	if err := h.repo.SoftDeleteFile(c.Request.Context(), id, userName); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "file not found or not owned"})
		return
	}

	// Publish file.deleted Kafka event (best-effort; log on failure but still return 200)
	event := models.FileDeletedEvent{FileID: id}
	payload, _ := json.Marshal(event)
	_ = h.kafkaWriter.WriteMessages(context.Background(), kafka.Message{Value: payload})

	c.JSON(http.StatusOK, gin.H{"success": true})
}
