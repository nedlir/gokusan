package main

import (
	"log"

	"auth/config"
	"auth/handlers"
	"auth/pkg/jwks"
	"auth/pkg/keycloak"
	"auth/routes"
	"auth/service/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	redisClient, err := config.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	kc := keycloak.New(cfg.KeycloakURL, cfg.KeycloakRealm, cfg.AdminUsername, cfg.AdminPassword)
	jwksURL := cfg.KeycloakURL + "/realms/" + cfg.KeycloakRealm + "/protocol/openid-connect/certs"
	jwksClient := jwks.NewJWKSClient(jwksURL, cfg.KeycloakRealm, cfg.JWKSCacheTTL, redisClient)

	svc := auth.NewService(cfg, kc, jwksClient)

	if err := svc.EnsureAdminUser(); err != nil {
		log.Printf("Warning: Could not ensure admin user: %v", err)
	}

	r := gin.Default()
	routes.Setup(r, cfg, handlers.NewAuthHandler(svc))

	log.Println("Auth service starting on :8080")
	log.Fatal(r.Run(":8080"))
}
