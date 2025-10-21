package respository

import (
	"context"
	"time"
)

type Auths struct {
	Username string `db:"username"`
	Password string `db:"password"`
}
type RefreshToken struct {
	Id        uint      `db:"id"`
	User_Id   int       `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	IsRevoked bool      `db:"is_revoked"`
}

type AccessToken struct {
	Id        uint      `db:"id"`
	User_Id   uint      `db:"user_id"`
	Token     string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	IsRevoked bool      `db:"is_revoked"`
	CreatedAt time.Time `db:"created_at"`
}
type Users struct {
	Id        uint      `db:"id"`
	Username  string    `db:"username"`
	FullName  string    `db:"full_name"`
	Password  string    `db:"password"`
	RoleID    int       `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Is_active bool      `db:"is_active"`
}

type AuthRepository interface {
	Login(ctx context.Context, username string) (*Users, error)
	GetUserByID(userID uint) (*Users, error)
	// RegisterUser(user Users) (*Users, error)
	FindValidRefreshTokenByHash(hashed string) (*RefreshToken, error)
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	RevokeToken(tokenID uint) error
	RevokeAllTokensForUser(userID uint) error
	RevokeAllTokens(userID uint) error
	RevokeAllTokensByUserID(userID uint) error
	RevokeExpiredTokensForAllUsers() error

	Save(ctx context.Context, token *AccessToken) error
	GetByToken(ctx context.Context, tokenString string) (*AccessToken, error)
	RevokeAllACCForUser(userID uint) error
	RevokeAllTokensACC(userID uint) error

	FindActiveAccessToken(token string) (*AccessToken, error)

	FindValidRefreshTokenByHashACC(hashed string) (*AccessToken, error)
	RevokeTokenACC(tokenID uint) error
}
