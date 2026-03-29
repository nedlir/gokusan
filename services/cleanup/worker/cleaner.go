package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
)

// File mirrors the relevant fields from the Metadata Service response.
type File struct {
	ID         string `json:"id"`
	StorageKey string `json:"storageKey"`
	Status     string `json:"status"`
}

type filesResponse struct {
	Success bool   `json:"success"`
	Files   []File `json:"files"`
}

type fileDeletedEvent struct {
	FileID string `json:"fileId"`
}

// Cleaner holds dependencies for the cleanup cron job.
type Cleaner struct {
	mc                 *minio.Client
	kafkaWriter        *kafka.Writer
	metadataServiceURL string
	bucket             string
	httpClient         *http.Client
}

// New creates a Cleaner.
func New(mc *minio.Client, kw *kafka.Writer, metadataURL, bucket string) *Cleaner {
	return &Cleaner{
		mc:                 mc,
		kafkaWriter:        kw,
		metadataServiceURL: metadataURL,
		bucket:             bucket,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
	}
}

// Run is the cron handler — fetches quarantined/deleted files and cleans them up.
func (c *Cleaner) Run() {
	log.Println("[cleaner] starting cleanup run")

	files, err := c.fetchFiles()
	if err != nil {
		log.Printf("[cleaner] failed to fetch files from metadata service: %v", err)
		return
	}

	log.Printf("[cleaner] found %d file(s) to clean", len(files))

	for _, f := range files {
		if err := c.cleanFile(f); err != nil {
			log.Printf("[cleaner] failed to clean file %s: %v", f.ID, err)
		}
	}

	log.Println("[cleaner] cleanup run complete")
}

// fetchFiles queries the Metadata Service for quarantined and deleted files.
func (c *Cleaner) fetchFiles() ([]File, error) {
	url := fmt.Sprintf("%s/files?status=quarantined,deleted", c.metadataServiceURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("metadata service returned %d: %s", resp.StatusCode, body)
	}

	var result filesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Files, nil
}

// cleanFile deletes the MinIO object and publishes a file.deleted Kafka event.
func (c *Cleaner) cleanFile(f File) error {
	if f.StorageKey != "" {
		err := c.mc.RemoveObject(context.Background(), c.bucket, f.StorageKey, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("remove minio object %s: %w", f.StorageKey, err)
		}
		log.Printf("[cleaner] deleted MinIO object %s for file %s", f.StorageKey, f.ID)
	}

	event := fileDeletedEvent{FileID: f.ID}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := c.kafkaWriter.WriteMessages(context.Background(), kafka.Message{Value: payload}); err != nil {
		return fmt.Errorf("publish file.deleted event for %s: %w", f.ID, err)
	}

	log.Printf("[cleaner] published file.deleted event for file %s", f.ID)
	return nil
}
