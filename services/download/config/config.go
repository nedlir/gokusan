package config

import "os"

type Config struct {
	Port               string
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucket        string
	MinIOUseSSL        bool
	MetadataServiceURL string
}

func Load() *Config {
	useSSL := false
	if v := os.Getenv("MINIO_USE_SSL"); v == "true" {
		useSSL = true
	}
	return &Config{
		Port:               getEnv("PORT", "8012"),
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:        getEnv("MINIO_BUCKET", "gokusan"),
		MinIOUseSSL:        useSSL,
		MetadataServiceURL: getEnv("METADATA_SERVICE_URL", "http://metadata:8013"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
