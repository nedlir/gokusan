package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	minio "github.com/minio/minio-go/v7"
	"download/models"
)

// Handler holds dependencies for the download handler.
type Handler struct {
	minioClient        *minio.Client
	bucket             string
	metadataServiceURL string
}

// New creates a Handler with the given dependencies.
func New(mc *minio.Client, bucket, metadataServiceURL string) *Handler {
	return &Handler{
		minioClient:        mc,
		bucket:             bucket,
		metadataServiceURL: metadataServiceURL,
	}
}

// Download handles GET /download/:id.
func (h *Handler) Download(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Success: false, Error: "Unauthorized"})
		return
	}

	fileID := c.Param("id")

	// Call Metadata Service to get file info.
	metaURL := fmt.Sprintf("%s/files/%s", h.metadataServiceURL, fileID)
	resp, err := http.Get(metaURL) //nolint:noctx
	if err != nil {
		log.Printf("metadata service request failed for %s: %v", fileID, err)
		c.JSON(http.StatusConflict, models.ErrorResponse{Success: false, Error: "File not available"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusConflict, models.ErrorResponse{Success: false, Error: "File not available"})
		return
	}

	var metaResp models.MetadataFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&metaResp); err != nil {
		log.Printf("failed to decode metadata response for %s: %v", fileID, err)
		c.JSON(http.StatusConflict, models.ErrorResponse{Success: false, Error: "File not available"})
		return
	}

	file := metaResp.File

	if file.Status != "ready" {
		c.JSON(http.StatusConflict, models.ErrorResponse{Success: false, Error: "File not available"})
		return
	}

	if file.OwnerID != userName {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Success: false, Error: "Forbidden"})
		return
	}

	// Stream from MinIO.
	storageKey := fmt.Sprintf("clean/%s", fileID)
	obj, err := h.minioClient.GetObject(c.Request.Context(), h.bucket, storageKey, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("MinIO GetObject failed for %s: %v", storageKey, err)
		c.JSON(http.StatusBadGateway, models.ErrorResponse{Success: false, Error: "Storage read failed"})
		return
	}
	defer obj.Close()

	mimeType := file.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	c.Header("Content-Type", mimeType)

	if _, err := io.Copy(c.Writer, obj); err != nil {
		log.Printf("streaming failed for %s: %v", storageKey, err)
		// Headers already sent; can't change status code at this point.
	}
}
