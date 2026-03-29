package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/gokusan/metadata/config"
	"github.com/gokusan/metadata/consumer"
	"github.com/gokusan/metadata/handlers"
	"github.com/gokusan/metadata/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("connected to database")

	repo := repository.New(pool)

	// Cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		cancel()
	}()

	// Start Kafka consumer
	cons := consumer.New(cfg.KafkaBroker, cfg.KafkaGroup, repo)
	go cons.Start(ctx)

	// Set up Gin router
	r := gin.Default()

	h := handlers.New(repo, cfg.KafkaBroker)

	r.GET("/health", h.Health)
	r.GET("/files", h.ListFiles)
	r.GET("/files/:id", h.GetFile)
	r.DELETE("/files/:id", h.DeleteFile)

	log.Printf("metadata service starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
