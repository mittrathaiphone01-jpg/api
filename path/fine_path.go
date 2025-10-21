package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func FinePath(app *fiber.App, h handler.FineRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/fines")
	v1 := api.Group("/v1")
	// private := v1.Group("/", middleware.RequireBillAuth())
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1), h.GetFines)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetFineById)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateFine)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateFine)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteFine)
}
