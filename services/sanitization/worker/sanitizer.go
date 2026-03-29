package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"

	"github.com/gokusan/sanitization/config"
)

// Event structs

type FileUploadedEvent struct {
	FileID     string `json:"fileId"`
	OwnerID    string `json:"ownerId"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	MimeType   string `json:"mimeType"`
	StorageKey string `json:"storageKey"`
}

type FileSanitizedEvent struct {
	FileID     string `json:"fileId"`
	StorageKey string `json:"storageKey"`
}

type FileQuarantinedEvent struct {
	FileID string `json:"fileId"`
}

// Worker holds all dependencies.
type Worker struct {
	mc     *minio.Client
	reader *kafka.Reader
	writer *kafka.Writer
	cfg    *config.Config
}

// New creates a new Worker.
func New(mc *minio.Client, reader *kafka.Reader, writer *kafka.Writer, cfg *config.Config) *Worker {
	return &Worker{mc: mc, reader: reader, writer: writer, cfg: cfg}
}

// Run starts the consume loop, blocking until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("[sanitizer] context cancelled, shutting down")
				return
			}
			log.Printf("[sanitizer] fetch error: %v", err)
			continue
		}

		w.processMessage(ctx, msg)
	}
}

func (w *Worker) processMessage(ctx context.Context, msg kafka.Message) {
	var event FileUploadedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[sanitizer] failed to unmarshal message: %v", err)
		// Commit to avoid infinite reprocessing of unparseable messages
		w.commitMessage(ctx, msg)
		return
	}

	log.Printf("[sanitizer] processing fileId=%s", event.FileID)

	tmpDir, err := os.MkdirTemp("", "sanitize-*")
	if err != nil {
		log.Printf("[sanitizer] fileId=%s failed to create temp dir: %v", event.FileID, err)
		w.commitMessage(ctx, msg)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Download raw file from MinIO
	rawKey := fmt.Sprintf("raw/%s", event.FileID)
	tmpFile := filepath.Join(tmpDir, event.FileName)

	if err := w.mc.FGetObject(context.Background(), w.cfg.MinIOBucket, rawKey, tmpFile, minio.GetObjectOptions{}); err != nil {
		log.Printf("[sanitizer] fileId=%s failed to download from MinIO: %v", event.FileID, err)
		w.commitMessage(ctx, msg)
		return
	}

	// Run Dangerzone with retry
	clean, err := w.runDangerzone(tmpDir, event.FileName)
	if err != nil {
		// Container invocation failed after retries — send to DLQ
		log.Printf("[sanitizer] fileId=%s dangerzone invocation failed after retries: %v", event.FileID, err)
		w.publishToDLQ(ctx, msg)
		w.commitMessage(ctx, msg)
		return
	}

	if clean {
		w.handleClean(ctx, event)
	} else {
		w.handleThreat(ctx, event)
	}

	w.commitMessage(ctx, msg)
}

// runDangerzone invokes the Dangerzone Docker container.
// Returns (true, nil) if clean, (false, nil) if threat detected, (false, err) if invocation failed.
func (w *Worker) runDangerzone(tmpDir, fileName string) (bool, error) {
	backoffs := []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		if attempt > 0 {
			time.Sleep(backoffs[attempt-1])
		}

		exitCode, err := w.execDocker(tmpDir, fileName)
		if err == nil {
			// Command ran successfully; interpret exit code
			return exitCode == 0, nil
		}

		// err != nil means the container invocation itself failed (not a threat result)
		lastErr = err
		log.Printf("[sanitizer] dangerzone attempt %d/%d failed: %v", attempt+1, len(backoffs)+1, err)
	}

	return false, fmt.Errorf("dangerzone failed after %d attempts: %w", len(backoffs)+1, lastErr)
}

// execDocker runs the Dangerzone container and returns the exit code.
// Returns an error only if the container could not be invoked (not for non-zero exit codes).
func (w *Worker) execDocker(tmpDir, fileName string) (int, error) {
	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", tmpDir+":/tmp",
		w.cfg.DangerzoneImage,
		fileName,
	)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Container ran but returned non-zero — threat detected, not an invocation failure
			return exitErr.ExitCode(), nil
		}
		// Invocation failure (Docker not available, image pull error, etc.)
		return -1, err
	}

	return 0, nil
}

func (w *Worker) handleClean(ctx context.Context, event FileUploadedEvent) {
	rawKey := fmt.Sprintf("raw/%s", event.FileID)
	cleanKey := fmt.Sprintf("clean/%s", event.FileID)

	// Copy sanitized file to clean/ prefix
	src := minio.CopySrcOptions{Bucket: w.cfg.MinIOBucket, Object: rawKey}
	dst := minio.CopyDestOptions{Bucket: w.cfg.MinIOBucket, Object: cleanKey}
	if _, err := w.mc.CopyObject(context.Background(), dst, src); err != nil {
		log.Printf("[sanitizer] fileId=%s failed to copy to clean/: %v", event.FileID, err)
		return
	}

	// Delete raw file
	if err := w.mc.RemoveObject(context.Background(), w.cfg.MinIOBucket, rawKey, minio.RemoveObjectOptions{}); err != nil {
		log.Printf("[sanitizer] fileId=%s failed to delete raw/: %v", event.FileID, err)
	}

	// Publish file.sanitized
	payload, _ := json.Marshal(FileSanitizedEvent{
		FileID:     event.FileID,
		StorageKey: cleanKey,
	})
	if err := w.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: "file.sanitized",
		Value: payload,
	}); err != nil {
		log.Printf("[sanitizer] fileId=%s failed to publish file.sanitized: %v", event.FileID, err)
	}

	log.Printf("[sanitizer] fileId=%s clean — published file.sanitized", event.FileID)
}

func (w *Worker) handleThreat(ctx context.Context, event FileUploadedEvent) {
	payload, _ := json.Marshal(FileQuarantinedEvent{FileID: event.FileID})
	if err := w.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: "file.quarantined",
		Value: payload,
	}); err != nil {
		log.Printf("[sanitizer] fileId=%s failed to publish file.quarantined: %v", event.FileID, err)
	}

	log.Printf("[sanitizer] fileId=%s threat detected — published file.quarantined", event.FileID)
}

func (w *Worker) publishToDLQ(ctx context.Context, msg kafka.Message) {
	if err := w.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: w.cfg.KafkaDLQTopic,
		Value: msg.Value,
	}); err != nil {
		log.Printf("[sanitizer] failed to publish to DLQ: %v", err)
	}
}

func (w *Worker) commitMessage(ctx context.Context, msg kafka.Message) {
	if err := w.reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("[sanitizer] failed to commit message: %v", err)
	}
}
