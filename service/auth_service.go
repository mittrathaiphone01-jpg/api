package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"rrmobile/respository"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// var limiter = rate.NewLimiter(rate.Every(time.Minute), 100) // 5 requests per minute

type authService struct {
	authRepository respository.AuthRepository
}

func NewAuthService(authRepository respository.AuthRepository) AuthService {
	return &authService{authRepository: authRepository}
}

func GenerateToken(userID uint, username string, roleID int) (string, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleId:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)), // ✅ อายุ 15 นาที
			ID:        uuid.NewString(),                             // ✅ jti unique
		},
	}

	// สร้าง token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte(viper.GetString("SECRET_KEY"))
	if len(secretKey) == 0 {
		return "", ErrJWTGenerationFailed
	}

	return token.SignedString(secretKey)
}
func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	// ตรวจสอบ JWT ก่อน
	secretKey := []byte(viper.GetString("SECRET_KEY"))
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	// ตรวจสอบหมดอายุ
	if claims.ExpiresAt == nil || now.After(claims.ExpiresAt.Time.Add(30*time.Second)) {
		return nil, errors.New("token expired")
	}

	// 🔑 Hash token ก่อนค้นหาใน DB
	hashed := HashToken(tokenString)
	acct, err := s.authRepository.FindActiveAccessToken(hashed)
	if err != nil || acct == nil || acct.IsRevoked {
		return nil, errors.New("token revoked")
	}

	return claims, nil
}

func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", ErrJWTGenerationFailed
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *authService) SaveRefreshToken(ctx context.Context, userID uint, token string, expires time.Time) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	if token == "" {
		return errors.New("token cannot be empty")
	}
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	hashed := HashToken(token)
	expires = now.Add(7 * 24 * time.Hour) // หมดอายุใน 1 นาที
	rt := respository.RefreshToken{
		User_Id:   int(userID),
		TokenHash: hashed,
		ExpiresAt: expires,
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	err := s.authRepository.SaveRefreshToken(ctx, &rt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: ", err)
	}

	return err
}
func (s *authService) SaveACCToken(ctx context.Context, userID uint, token string, expires time.Time) error {
	if userID == 0 || token == "" {
		return errors.New("invalid input")
	}

	hashed := HashToken(token) // SHA256
	acct := respository.AccessToken{
		User_Id:   userID,
		Token:     hashed,
		ExpiresAt: expires,
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	return s.authRepository.Save(ctx, &acct)
}
func (s *authService) RefreshTokens(ctx context.Context, oldToken string) (string, string, error) {
	if oldToken == "" {
		return "", "", errors.New("refresh token cannot be empty")
	}

	// แปลง token เป็น hash เพื่อค้นหาใน DB
	hashed := HashToken(oldToken)
	rt, err := s.authRepository.FindValidRefreshTokenByHash(hashed)
	if err != nil || rt == nil {
		return "", "", ErrTokenExpiredOrInvalid
	}

	// ใช้เวลาไทย
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	// เช็คหมดอายุ refresh token
	if now.After(rt.ExpiresAt.In(loc)) {
		_ = s.authRepository.RevokeToken(rt.Id)
		return "", "", errors.New("refresh token expired, please login again")
	}

	// ดึงข้อมูล user
	user, err := s.authRepository.GetUserByID(uint(rt.User_Id))
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve user information: ", err)
	}

	// Revoke access token เก่าทั้งหมด
	if err := s.authRepository.RevokeAllACCForUser(user.Id); err != nil {
		return "", "", fmt.Errorf("failed to revoke existing tokens: ", err)
	}

	// สร้าง access token ใหม่
	accessToken, err := GenerateToken(user.Id, user.Username, user.RoleID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: ", err)
	}

	// บันทึก access token ใหม่ (อายุ 15 นาที)
	if err := s.SaveACCToken(ctx, user.Id, accessToken, now.Add(5*time.Minute)); err != nil {
		return "", "", fmt.Errorf("failed to save access token: ", err)
	}

	return accessToken, oldToken, nil
}

func (s *authService) RevokeAllTokensForUser(userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	return s.authRepository.RevokeAllTokensForUser(userID)
}
func (s *authService) RevokeAllACCTokensForUser(userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	return s.authRepository.RevokeAllACCForUser(userID)
}
func (s *authService) LogoutAndRevokeAll(userID uint) error {
	// 1. Validate input
	if userID == 0 {
		return errors.New("invalid user ID: cannot be zero")
	}

	// 2. ตรวจสอบว่ามี user จริงในระบบ
	user, err := s.authRepository.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: ", err)
	}
	if user == nil {
		return fmt.Errorf("user with ID  not found", userID)
	}

	// 3. Revoke tokens ทั้งหมด
	if err := s.authRepository.RevokeAllTokens(userID); err != nil {
		return fmt.Errorf("failed to revoke tokens for user ", err)
	}
	if err := s.authRepository.RevokeAllTokensACC(userID); err != nil {
		return fmt.Errorf("failed to revoke tokens for user ", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, username string, password string) (*AuthResponse, error) {
	validate := validator.New()
	if err := validate.Var(username, "required"); err != nil {

		return nil, errors.New("username is required")
	}
	if err := validate.Var(password, "required"); err != nil {

		return nil, errors.New("password is required")
	}

	auth, err := s.authRepository.Login(ctx, username)
	if err != nil {

		return nil, fmt.Errorf("failed to find user:", err)
	}
	if auth == nil {

		return nil, errors.New("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password)); err != nil {

		return nil, errors.New("invalid password")
	}

	// Revoke token เดิม
	if err := s.authRepository.RevokeAllTokensForUser(auth.Id); err != nil {
		return nil, fmt.Errorf("failed to revoke existing tokens: ", err)
	}
	if err := s.authRepository.RevokeAllACCForUser(auth.Id); err != nil {
		return nil, fmt.Errorf("failed to revoke existing tokens: ", err)
	}

	accessToken, err := GenerateToken(auth.Id, auth.Username, auth.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	// ตั้งค่าเวลาไทย
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	// บันทึก access token ลง DB หมดอายุ 15 นาที (ตัวอย่าง)
	expiresAt_Acc := now.Add(5 * time.Minute)
	if err := s.SaveACCToken(ctx, auth.Id, accessToken, expiresAt_Acc); err != nil {
		return nil, fmt.Errorf("failed to save access token:", err)
	}

	refreshToken, err := GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: ", err)
	}

	// หมดอายุ refresh token 7 วัน
	expiresAt_Refresh := now.Add(7 * 24 * time.Hour)
	// expiresAt_Refresh := now.Add(5 * time.Minute)
	if err := s.SaveRefreshToken(ctx, auth.Id, refreshToken, expiresAt_Refresh); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: ", err)
	}

	return &AuthResponse{
		Id:           auth.Id,
		Username:     auth.Username,
		Token:        accessToken,
		RoleID:       auth.RoleID,
		RefreshToken: refreshToken,
	}, nil
}
