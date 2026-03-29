package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	KafkaBroker string
	KafkaGroup  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8013"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/gokusan?sslmode=disable"),
		KafkaBroker: getEnv("KAFKA_BROKER", "kafka:9092"),
		KafkaGroup:  getEnv("KAFKA_GROUP", "metadata-service"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
