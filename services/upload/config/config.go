package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	KafkaBroker    string
	MaxUploadBytes int64
}

func Load() *Config {
	maxBytes := int64(100 * 1024 * 1024) // 100 MB default
	if v := os.Getenv("MAX_UPLOAD_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxBytes = n
		}
	}
	useSSL := false
	if v := os.Getenv("MINIO_USE_SSL"); v == "true" {
		useSSL = true
	}
	return &Config{
		Port:           getEnv("PORT", "6565"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "gokusan"),
		MinIOUseSSL:    useSSL,
		KafkaBroker:    getEnv("KAFKA_BROKER", "kafka:9092"),
		MaxUploadBytes: maxBytes,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
