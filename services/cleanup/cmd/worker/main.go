package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/robfig/cron/v3"
	"github.com/segmentio/kafka-go"

	"github.com/gokusan/cleanup/config"
	"github.com/gokusan/cleanup/worker"
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

	// Kafka writer for file.deleted events
	kw := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kw.Close()

	cleaner := worker.New(mc, kw, cfg.MetadataServiceURL, cfg.MinIOBucket)

	c := cron.New()
	if _, err := c.AddFunc(cfg.CronSchedule, cleaner.Run); err != nil {
		log.Fatalf("[main] invalid cron schedule %q: %v", cfg.CronSchedule, err)
	}
	c.Start()
	log.Printf("[main] cleanup worker started with schedule %q", cfg.CronSchedule)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("[main] received signal %s, shutting down", sig)

	ctx := c.Stop()
	<-ctx.Done()
	log.Println("[main] cleanup worker stopped")
}
