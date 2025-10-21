package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthResponse struct {
	Id           uint   `json:"id"`
	Username     string `json:"username"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	RoleID       int    `json:"role_id"`
}

type AuthRegisterResponse struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type AuthLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type NewAuthRequest struct {
	Username        string `json:"username" validate:"required"`
	Password        string `json:"password" validate:"required"`
	FullName        string `json:"fullname" validate:"required"`
	ConfirmPassword string `json:"confirmpassword" validate:"required"`
	RoleID          int    `json:"role_id"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleId   int    `json:"role_id"`

	jwt.RegisteredClaims
}

type AuthService interface {
	Login(ctx context.Context, username string, password string) (*AuthResponse, error)
	SaveRefreshToken(ctx context.Context, userID uint, token string, expiresAt time.Time) error
	// Register(request NewAuthRequest) (*AuthRegisterResponse, error)
	RevokeAllTokensForUser(userID uint) error
	// RefreshTokens(ctx context.Context, oldToken string) (string, string, error)
	RefreshTokens(ctx context.Context, oldRefreshToken string) (string, string, error)
	LogoutAndRevokeAll(userID uint) error
	ValidateToken(tokenString string) (*Claims, error)
}
