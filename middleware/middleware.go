package middleware

import (
	"strings"

	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func RequireBillAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).SendFile("./static/404.html")

		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		expectedToken := strings.TrimSpace(viper.GetString("BILL_API_TOKEN"))

		if token != expectedToken {
			return c.Status(fiber.StatusUnauthorized).SendFile("./static/404.html")

		}

		return c.Next()
	}
}
func JWTMiddleware(authSvc service.AuthService, usersSvc service.UsersService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
		}

		claims, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		user, err := usersSvc.GetUserById(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "บัญชีไม่พบ"})
		}
		if !user.Is_active {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "บัญชีของคุณถูกปิดใช้งาน, กรุณาเข้าสู่ระบบใหม่"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role_id", claims.RoleId)

		return c.Next()
	}
}

func RoleMiddleware(authSvc service.AuthService, allowedRoles ...int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "กรุณาเข้าสู่ระบบ",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "กรุณาเข้าสู่ระบบ",
			})
		}

		claims, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 4. ตรวจสอบ role
		for _, role := range allowedRoles {
			if claims.RoleId == role {
				c.Locals("user_id", claims.UserID)
				c.Locals("username", claims.Username)
				c.Locals("role_id", claims.RoleId)
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "คุณไม่มีสิทธิ์เข้าถึงหน้านี้",
		})
	}
}
