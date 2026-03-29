package config

import "os"

type Config struct {
	CronSchedule       string
	MetadataServiceURL string
	KafkaBrokers       []string
	KafkaTopic         string
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucket        string
	MinIOUseSSL        bool
}

func Load() *Config {
	return &Config{
		CronSchedule:       getEnv("CRON_SCHEDULE", "@hourly"),
		MetadataServiceURL: getEnv("METADATA_SERVICE_URL", "http://metadata:8013"),
		KafkaBrokers:       []string{getEnv("KAFKA_BROKERS", "kafka:9092")},
		KafkaTopic:         getEnv("KAFKA_TOPIC", "file.deleted"),
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:        getEnv("MINIO_BUCKET", "gokusan"),
		MinIOUseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
