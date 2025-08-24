package keycloak

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

type Client interface {
	Login(ctx context.Context, username, password string) (accessToken string, err error)
	LoginAdmin(ctx context.Context) (accessToken string, err error)
	CreateUser(ctx context.Context, adminToken string, user gocloak.User) (string, error)
	GetRealmRole(ctx context.Context, adminToken, roleName string) (*gocloak.Role, error)
	AddRealmRoleToUser(ctx context.Context, adminToken, userID string, roles []gocloak.Role) error
}

type goCloakClient struct {
	kc            *gocloak.GoCloak
	realm         string
	adminUsername string
	adminPassword string
}

func New(url, realm, adminUsername, adminPassword string) Client {
	return &goCloakClient{
		kc:            gocloak.NewClient(url),
		realm:         realm,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}
}

func (c *goCloakClient) Login(ctx context.Context, username, password string) (string, error) {
	token, err := c.kc.Login(ctx, "admin-cli", "", c.realm, username, password)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (c *goCloakClient) LoginAdmin(ctx context.Context) (string, error) {
	token, err := c.kc.LoginAdmin(ctx, c.adminUsername, c.adminPassword, c.realm)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (c *goCloakClient) CreateUser(ctx context.Context, adminToken string, user gocloak.User) (string, error) {
	return c.kc.CreateUser(ctx, adminToken, c.realm, user)
}

func (c *goCloakClient) GetRealmRole(ctx context.Context, adminToken, roleName string) (*gocloak.Role, error) {
	return c.kc.GetRealmRole(ctx, adminToken, c.realm, roleName)
}

func (c *goCloakClient) AddRealmRoleToUser(ctx context.Context, adminToken, userID string, roles []gocloak.Role) error {
	return c.kc.AddRealmRoleToUser(ctx, adminToken, c.realm, userID, roles)
}
