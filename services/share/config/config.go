package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	MetadataServiceURL string
	DownloadServiceURL string
}

func Load() *Config {
	redisDB := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			redisDB = n
		}
	}
	return &Config{
		Port:               getEnv("PORT", "8014"),
		RedisAddr:          getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		MetadataServiceURL: getEnv("METADATA_SERVICE_URL", "http://metadata:8013"),
		DownloadServiceURL: getEnv("DOWNLOAD_SERVICE_URL", "http://download:8012"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
