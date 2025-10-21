package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)


func GenerateImageToken(filename string) (string, error) {
	// ตั้งเวลา token หมดอายุ 1 นาที
	expiration := time.Now().Add(1 * time.Minute).Unix()

	claims := jwt.MapClaims{
		"filename": filename,
		"exp":      expiration,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := strings.TrimSpace(viper.GetString("SECRET_KEY"))
	if secret == "" {
		return "", fmt.Errorf("SECRET_KEY is not set")
	}

	return token.SignedString([]byte(secret))
}
