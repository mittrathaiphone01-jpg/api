package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func InstallmentPath(app *fiber.App, h handler.InstallmentRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/installments")
	v1 := api.Group("/v1")
	// private := v1.Group("/", middleware.RequireBillAuth())
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1, 2), h.GetInstallments)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetInstallmentById)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateInstallment)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateInstallment)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteInstallment)
}
