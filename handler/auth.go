package handler

import (
	"rrmobile/model"
	"rrmobile/service"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthRequestHandler interface {
	Login(c *fiber.Ctx) error
	// Register(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

type authHandler struct {
	authService service.AuthService
	db          *gorm.DB
}

func NewAuthHandler(authService service.AuthService, db *gorm.DB) *authHandler {
	return &authHandler{authService: authService, db: db}
}

func (ah *authHandler) Login(c *fiber.Ctx) error {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ส่งข้อมูลมาไม่ครบ"})
	}

	authResponse, err := ah.authService.Login(c.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "username or password ไม่ถูกต้อง"})
	}

	return c.JSON(fiber.Map{
		"role_id":       authResponse.RoleID,
		"access_token":  authResponse.Token,
		"refresh_token": authResponse.RefreshToken,
	})
}

func (ah *authHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ส่งข้อมูลมาไม่ครบ"})
	}
	newAccessToken, newRefreshToken, err := ah.authService.RefreshTokens(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ข้อมูล ไม่ถูกต้อง หรือ หมดอายุ"})
	}

	return c.JSON(fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func GetUsersHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var users []model.Users
		if err := db.Find(&users).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(users)
	}
}

func (h *authHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header missing",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID := claims.UserID

	if err := h.authService.LogoutAndRevokeAll(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User logged out and all tokens revoked",
	})
}
