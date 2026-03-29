package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gokusan/metadata/models"
	"github.com/gokusan/metadata/repository"
	"github.com/segmentio/kafka-go"
)

const (
	maxRetries = 3
	retryDelay = time.Second
)

var topics = []string{
	"file.uploaded",
	"file.sanitized",
	"file.quarantined",
	"file.deleted",
}

type Consumer struct {
	broker string
	group  string
	repo   *repository.Repository
}

func New(broker, group string, repo *repository.Repository) *Consumer {
	return &Consumer{broker: broker, group: group, repo: repo}
}

func (c *Consumer) Start(ctx context.Context) {
	for _, topic := range topics {
		go c.consume(ctx, topic)
	}
	<-ctx.Done()
}

func (c *Consumer) consume(ctx context.Context, topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{c.broker},
		Topic:    topic,
		GroupID:  c.group,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[consumer] fetch error on %s: %v", topic, err)
			continue
		}

		var processErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			processErr = c.handle(ctx, topic, msg)
			if processErr == nil {
				break
			}
			log.Printf("[consumer] attempt %d/%d failed for topic %s: %v", attempt, maxRetries, topic, processErr)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
			}
		}

		if processErr != nil {
			log.Printf("[consumer] sending to DLQ after %d failures: topic=%s", maxRetries, topic)
			c.publishDLQ(topic, msg)
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			log.Printf("[consumer] commit error on %s: %v", topic, err)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, topic string, msg kafka.Message) error {
	switch topic {
	case "file.uploaded":
		var e models.FileUploadedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return err
		}
		return c.repo.InsertFile(ctx, &models.File{
			ID:         e.FileID,
			OwnerID:    e.OwnerID,
			Name:       e.FileName,
			Size:       e.FileSize,
			MimeType:   e.MimeType,
			StorageKey: e.StorageKey,
		})

	case "file.sanitized":
		var e models.FileSanitizedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return err
		}
		return c.repo.UpdateFileStatus(ctx, e.FileID, "ready", e.StorageKey)

	case "file.quarantined":
		var e models.FileQuarantinedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return err
		}
		return c.repo.UpdateFileStatus(ctx, e.FileID, "quarantined", "")

	case "file.deleted":
		var e models.FileDeletedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return err
		}
		return c.repo.UpdateFileStatus(ctx, e.FileID, "deleted", "")
	}
	return nil
}

func (c *Consumer) publishDLQ(topic string, msg kafka.Message) {
	dlqTopic := topic + ".dlq"
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{c.broker},
		Topic:   dlqTopic,
	})
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := w.WriteMessages(ctx, kafka.Message{Value: msg.Value}); err != nil {
		log.Printf("[consumer] failed to publish to DLQ %s: %v", dlqTopic, err)
	}
}
