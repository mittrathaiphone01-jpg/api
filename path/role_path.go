package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func RolesPath(app *fiber.App, h handler.RolesRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/roles")
	v1 := api.Group("/v1")
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

	// v1.Use(middleware.MiddlewareJWT(authSvc))     // ✅ ส่ง authSvc
	// v1.Use(middleware.AuthMiddleware(authSvc, 1)) // ✅ ส่ง authSvc ก่อน แล้ว roleId

	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1), h.GetAllRoles)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetRoleByID)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateRole)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteRole)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateRoles)
}
