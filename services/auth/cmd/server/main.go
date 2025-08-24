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

	kc := keycloak.New(cfg.KeycloakURL, cfg.KeycloakRealm, cfg.AdminUsername, cfg.AdminPassword)
	jwksClient := jwks.NewJWKSClient(cfg.KeycloakURL+"/realms/"+cfg.KeycloakRealm+"/protocol/openid-connect/certs", cfg.JWKSCacheTTL)

	svc := auth.NewService(cfg, kc, jwksClient)

	if err := svc.EnsureAdminUser(); err != nil {
		log.Printf("Warning: Could not ensure admin user: %v", err)
	}

	r := gin.Default()
	routes.Setup(r, cfg, handlers.NewAuthHandler(svc))

	log.Println("Auth service starting on :8080")
	log.Fatal(r.Run(":8080"))
}
