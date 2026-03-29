package config

import "os"

type Config struct {
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaDLQTopic   string
	KafkaGroupID    string
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOBucket     string
	MinIOUseSSL     bool
	DangerzoneImage string
}

func Load() *Config {
	return &Config{
		KafkaBrokers:    []string{getEnv("KAFKA_BROKERS", "kafka:9092")},
		KafkaTopic:      getEnv("KAFKA_TOPIC", "file.uploaded"),
		KafkaDLQTopic:   getEnv("KAFKA_DLQ_TOPIC", "file.uploaded.dlq"),
		KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "sanitization-worker"),
		MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:     getEnv("MINIO_BUCKET", "files"),
		MinIOUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		DangerzoneImage: getEnv("DANGERZONE_IMAGE", "ghcr.io/freedomofpress/dangerzone/v1"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
