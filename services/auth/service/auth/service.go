package auth

import (
	"context"
	"fmt"
	"log"
	"strings"

	"auth/config"
	"auth/models"
	"auth/pkg/jwks"
	"auth/pkg/keycloak"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v4"
)

type Service struct {
	cfg        config.Config
	kc         keycloak.Client
	jwksClient *jwks.JWKSClient
}

func NewService(cfg config.Config, kc keycloak.Client, jwksClient *jwks.JWKSClient) *Service {
	return &Service{cfg: cfg, kc: kc, jwksClient: jwksClient}
}

func (s *Service) EnsureAdminUser() error {
	ctx := context.Background()
	if _, err := s.kc.Login(ctx, "admin", "admin"); err == nil {
		log.Printf("Admin user already exists and is functional")
		return nil
	}
	log.Printf("Admin user doesn't exist, creating...")
	adminToken, err := s.kc.LoginAdmin(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %v", err)
	}
	user := gocloak.User{
		Username: gocloak.StringP("admin"),
		Email:    gocloak.StringP("admin@gokusan.local"),
		Enabled:  gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP("admin"),
			Temporary: gocloak.BoolP(false),
		}},
	}
	userID, err := s.kc.CreateUser(ctx, adminToken, user)
	if err != nil && !strings.Contains(err.Error(), "409") {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	adminRole, err := s.kc.GetRealmRole(ctx, adminToken, "admin")
	if err == nil {
		_ = s.kc.AddRealmRoleToUser(ctx, adminToken, userID, []gocloak.Role{*adminRole})
	}
	return nil
}

func (s *Service) Authenticate(username, password string) (string, models.User, error) {
	log.Printf("Attempting Keycloak authentication for user: %s", username)
	token, err := s.kc.Login(context.Background(), username, password)
	if err != nil {
		return "", models.User{}, fmt.Errorf("authentication failed")
	}
	user, err := s.ParseAndValidateToken(token)
	if err != nil {
		return "", models.User{}, fmt.Errorf("failed to extract user info: %v", err)
	}
	return token, user, nil
}

func (s *Service) Register(username, email, password string) error {
	ctx := context.Background()
	adminToken, err := s.kc.LoginAdmin(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %v", err)
	}
	user := gocloak.User{
		Username: gocloak.StringP(username),
		Email:    gocloak.StringP(email),
		Enabled:  gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(password),
			Temporary: gocloak.BoolP(false),
		}},
	}
	_, err = s.kc.CreateUser(ctx, adminToken, user)
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

func (s *Service) ParseAndValidateToken(tokenString string) (models.User, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid found in token header")
		}
		return s.jwksClient.GetKey(kid)
	})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to parse token: %v", err)
	}
	if !parsedToken.Valid {
		return models.User{}, fmt.Errorf("token is not valid")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return models.User{}, fmt.Errorf("invalid token claims")
	}
	return s.extractUserFromClaims(claims)
}

func (s *Service) extractUserFromClaims(claims jwt.MapClaims) (models.User, error) {
	if issuer, ok := claims["iss"].(string); !ok {
		return models.User{}, fmt.Errorf("missing token issuer")
	} else if !strings.Contains(issuer, "/realms/"+s.cfg.KeycloakRealm) {
		return models.User{}, fmt.Errorf("invalid token issuer: %s", issuer)
	}

	var username string
	if preferred, ok := claims["preferred_username"].(string); ok {
		username = preferred
	} else if name, ok := claims["name"].(string); ok {
		username = name
	} else if sub, ok := claims["sub"].(string); ok {
		username = sub
	} else {
		return models.User{}, fmt.Errorf("no username found in token")
	}

	role := "user"
	if username == "admin" {
		role = "admin"
	} else if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, r := range roles {
				if roleStr, ok := r.(string); ok {
					if roleStr == "admin" || roleStr == "realm-admin" || roleStr == "create-realm" {
						role = "admin"
						break
					}
				}
			}
		}
	}

	return models.User{Name: username, Role: role}, nil
}
