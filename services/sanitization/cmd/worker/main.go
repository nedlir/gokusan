package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"

	"github.com/gokusan/sanitization/config"
	"github.com/gokusan/sanitization/worker"
)

func main() {
	cfg := config.Load()

	// MinIO client
	mc, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatalf("[main] failed to create MinIO client: %v", err)
	}

	// Kafka reader (consumer group)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopic,
		GroupID:  cfg.KafkaGroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Kafka writer — no fixed topic; topic is set per message
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Graceful shutdown via OS signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("[main] received signal %s, shutting down", sig)
		cancel()
	}()

	log.Println("[main] sanitization worker started")

	w := worker.New(mc, reader, writer, cfg)
	w.Run(ctx)

	log.Println("[main] sanitization worker stopped")
}
