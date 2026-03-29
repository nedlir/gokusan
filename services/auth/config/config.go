package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	KeycloakURL   string
	KeycloakRealm string

	AdminUsername string
	AdminPassword string

	AllowedOrigin string

	CookieName      string
	CookieMaxAgeSec int
	JWKSCacheTTL    time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() Config {
	return Config{
		KeycloakURL:     getEnv("KEYCLOAK_URL", "http://keycloak:8080"),
		KeycloakRealm:   getEnv("KEYCLOAK_REALM", "master"),
		AdminUsername:   getEnv("KEYCLOAK_ADMIN_USERNAME", "admin"),
		AdminPassword:   getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin"),
		AllowedOrigin:   getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		CookieName:      getEnv("COOKIE_NAME", "auth_token"),
		CookieMaxAgeSec: getEnvInt("COOKIE_MAX_AGE_SEC", 24*60*60),
		JWKSCacheTTL:    getEnvDuration("JWKS_CACHE_TTL", 5*time.Minute),
		RedisAddr:       getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("REDIS_DB", 0),
	}
}

// NewRedisClient creates and pings a Redis client from the loaded config.
func NewRedisClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	log.Printf("Connected to Redis at %s", cfg.RedisAddr)
	return client, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
