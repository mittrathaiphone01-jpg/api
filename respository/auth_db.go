package respository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type authRepositoryDB struct {
	db *gorm.DB
}

func NewAuthRepositoryDB(db *gorm.DB) AuthRepository {
	return &authRepositoryDB{db: db}
}
func (r *authRepositoryDB) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
func (r *authRepositoryDB) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepositoryDB) RevokeAllTokensByUserID(userID uint) error {
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}
func (r *authRepositoryDB) RevokeExpiredTokensForAllUsers() error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("expires_at <= ? AND is_revoked = false", now).
		Update("is_revoked", true).Error
}

func (r *authRepositoryDB) RevokeAllTokensForUser(userID uint) error {
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}

func (r *authRepositoryDB) FindValidRefreshTokenByHash(hashed string) (*RefreshToken, error) {

	var token RefreshToken
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctx).Where("token_hash = ?", hashed).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find refresh token:", err)
	}

	if token.IsRevoked || time.Now().After(token.ExpiresAt) {
		_ = r.RevokeToken(token.Id)
		return nil, nil
	}

	expiration := token.ExpiresAt.Sub(time.Now())
	if expiration <= 0 {
		expiration = time.Second
	}

	return &token, nil
}
func (r *authRepositoryDB) FindValidRefreshTokenByHashACC(hashed string) (*AccessToken, error) {

	var token AccessToken
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctx).Where("token = ?", hashed).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find refresh token: ", err)
	}

	if token.IsRevoked || time.Now().After(token.ExpiresAt) {
		_ = r.RevokeTokenACC(token.Id)
		return nil, nil
	}

	expiration := token.ExpiresAt.Sub(time.Now())
	if expiration <= 0 {
		expiration = time.Second
	}

	return &token, nil
}

func (r *authRepositoryDB) RevokeToken(tokenID uint) error {
	var token RefreshToken
	if err := r.db.Where("id = ?", tokenID).First(&token).Error; err != nil {
		return fmt.Errorf("failed to find token with ID", err)
	}

	return r.db.Model(&RefreshToken{}).Where("id = ?", tokenID).Update("is_revoked", true).Error
}

func (r *authRepositoryDB) RevokeTokenACC(tokenID uint) error {
	var token AccessToken
	if err := r.db.Where("id = ?", tokenID).First(&token).Error; err != nil {
		return fmt.Errorf("failed to find token with ID ", err)
	}

	return r.db.Model(&RefreshToken{}).Where("id = ?", tokenID).Update("is_revoked", true).Error
}
func (r *authRepositoryDB) RevokeAllTokens(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&RefreshToken{}).Error
}

func (r *authRepositoryDB) Login(ctx context.Context, username string) (*Users, error) {
	var user Users

	// ตรวจสอบอินพุตก่อนใช้ใน query (แม้ GORM จะ escape ให้อยู่แล้ว)
	if username == "" {
		return nil, errors.New("username is required")
	}

	err := r.db.WithContext(ctx).
		Select("id", "username", "password", "role_id", "is_active").
		Where("username = ? AND is_active = ?", username, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ไม่เปิดเผยว่าผู้ใช้ไม่มีอยู่ หรือ inactive
			return nil, errors.New("invalid login credentials")
		}
		// log actual error (อย่า return error จริงให้ client)
		return nil, fmt.Errorf("database error:", err)
	}

	return &user, nil
}

func (r *authRepositoryDB) GetUserByID(userID uint) (*Users, error) {

	var user Users
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user:", err)
	}

	return &user, nil
}

func (r *authRepositoryDB) Save(ctx context.Context, token *AccessToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepositoryDB) GetByToken(ctx context.Context, tokenString string) (*AccessToken, error) {
	var token AccessToken
	if err := r.db.WithContext(ctx).Where("token = ?", tokenString).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *authRepositoryDB) RevokeAllACCForUser(userID uint) error {
	return r.db.Model(&AccessToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Update("is_revoked", true).Error
}

func (r *authRepositoryDB) FindActiveAccessToken(token string) (*AccessToken, error) {
	var at AccessToken
	err := r.db.
		Where("token = ? AND is_revoked = ?", token, false).
		First(&at).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// return nil, nil // ✅ ไม่ถือว่าเป็น error จริง
			return nil, nil
		}
		return nil, err
	}

	return &at, nil
}

func (r *authRepositoryDB) RevokeAllTokensACC(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&AccessToken{}).Error
}
