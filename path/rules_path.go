package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func RulesPath(app *fiber.App, h handler.RulesRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/rules")
	v1 := api.Group("/v1")
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))
	// v1.Use(middleware.MiddlewareJWT())
	// v1.Use(middleware.AuthMiddleware(1))
	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1), h.GetAllRules)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetRuleByID)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateRule)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.EditRule)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteRule)
}
