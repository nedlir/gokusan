package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gokusan/share/models"
	"github.com/redis/go-redis/v9"
)

// Handler holds dependencies for the share service handlers.
type Handler struct {
	redis              *redis.Client
	httpClient         *http.Client
	metadataServiceURL string
	downloadServiceURL string
}

// New creates a Handler with the given dependencies.
func New(rdb *redis.Client, httpClient *http.Client, metadataServiceURL, downloadServiceURL string) *Handler {
	return &Handler{
		redis:              rdb,
		httpClient:         httpClient,
		metadataServiceURL: metadataServiceURL,
		downloadServiceURL: downloadServiceURL,
	}
}

// CreateShare handles POST /share.
func (h *Handler) CreateShare(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Success: false, Error: "Unauthorized"})
		return
	}

	var req models.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	// Verify file via Metadata Service
	metaURL := fmt.Sprintf("%s/files/%s", h.metadataServiceURL, req.FileID)
	metaReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, metaURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}
	metaReq.Header.Set("X-User-Name", userName)

	resp, err := h.httpClient.Do(metaReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "File not found"})
		return
	}
	defer resp.Body.Close()

	var metaResp models.MetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&metaResp); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	if metaResp.File.Status != "ready" {
		c.JSON(http.StatusConflict, models.ErrorResponse{Success: false, Error: "File not available"})
		return
	}

	if metaResp.File.OwnerID != userName {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Success: false, Error: "Forbidden"})
		return
	}

	// Generate 32-byte crypto/rand token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Store in Redis
	record := models.ShareRecord{FileID: req.FileID, CreatedBy: userName}
	payload, err := json.Marshal(record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	key := "share:" + token
	ttl := time.Duration(req.TTL) * time.Second
	if err := h.redis.Set(context.Background(), key, payload, ttl).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "url": "/share/" + token})
}

// ResolveShare handles GET /share/:token — looks up the token and proxies the download.
func (h *Handler) ResolveShare(c *gin.Context) {
	token := c.Param("token")
	key := "share:" + token

	val, err := h.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Share link not found or expired"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	var record models.ShareRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	// Proxy download from Download Service
	dlURL := fmt.Sprintf("%s/download/%s", h.downloadServiceURL, record.FileID)
	dlReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, dlURL, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse{Success: false, Error: "Download failed"})
		return
	}

	dlResp, err := h.httpClient.Do(dlReq)
	if err != nil || dlResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, models.ErrorResponse{Success: false, Error: "Download failed"})
		return
	}
	defer dlResp.Body.Close()

	// Pass through relevant headers
	for _, header := range []string{"Content-Type", "Content-Disposition", "Content-Length"} {
		if v := dlResp.Header.Get(header); v != "" {
			c.Header(header, v)
		}
	}

	c.Status(http.StatusOK)
	io.Copy(c.Writer, dlResp.Body) //nolint:errcheck
}

// DeleteShare handles DELETE /share/:token.
func (h *Handler) DeleteShare(c *gin.Context) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Success: false, Error: "Unauthorized"})
		return
	}

	token := c.Param("token")
	key := "share:" + token

	val, err := h.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Share link not found or expired"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	var record models.ShareRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	if record.CreatedBy != userName {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Success: false, Error: "Forbidden"})
		return
	}

	if err := h.redis.Del(context.Background(), key).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
