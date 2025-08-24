package handlers

import (
	"log"
	"net/http"

	"auth/config"
	"auth/constants"
	"auth/models"
	"auth/service/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg config.Config
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) WithConfig(cfg config.Config) *AuthHandler {
	h.cfg = cfg
	return h
}

func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *AuthHandler) KongValidate(c *gin.Context) {
	log.Printf("Kong validation request from %s", c.ClientIP())
	token, err := c.Cookie(h.cfg.CookieName)
	if err != nil || token == "" {
		log.Printf("Kong validation failed: no auth cookie")
		c.Status(http.StatusUnauthorized)
		return
	}
	user, err := h.svc.ParseAndValidateToken(token)
	if err != nil {
		log.Printf("Kong validation failed: %v", err)
		c.Status(http.StatusUnauthorized)
		return
	}
	c.Header(constants.HeaderUserName, user.Name)
	c.Header(constants.HeaderUserRole, user.Role)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.LoginResponse{Success: false, Error: "Invalid request body"})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, models.LoginResponse{Success: false, Error: "Username and password are required"})
		return
	}
	token, user, err := h.svc.Authenticate(req.Username, req.Password)
	if err != nil {
		log.Printf("Authentication failed for user %s: %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, models.LoginResponse{Success: false, Error: "Invalid credentials"})
		return
	}
	c.SetCookie(h.cfg.CookieName, token, h.cfg.CookieMaxAgeSec, "/", "", false, true)
	c.JSON(http.StatusOK, models.LoginResponse{Success: true, User: user})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.RegisterResponse{Success: false, Error: "Invalid request body"})
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, models.RegisterResponse{Success: false, Error: "Username, email and password are required"})
		return
	}
	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, models.RegisterResponse{Success: false, Error: "Passwords do not match"})
		return
	}
	if req.Username == "admin" {
		c.JSON(http.StatusBadRequest, models.RegisterResponse{Success: false, Error: "Username 'admin' is reserved"})
		return
	}
	if err := h.svc.Register(req.Username, req.Email, req.Password); err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusInternalServerError, models.RegisterResponse{Success: false, Error: "Username already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, models.RegisterResponse{Success: false, Error: "Registration failed. Please try again."})
		}
		return
	}
	c.JSON(http.StatusOK, models.RegisterResponse{Success: true, Message: "User registered successfully! You can now login with your credentials."})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(h.cfg.CookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logged out successfully"})
}
