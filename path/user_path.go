package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func UsersPath(app *fiber.App, h handler.UsersRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	app.Get("test", h.GetAllUsers)

	api := app.Group("/users")
	v1 := api.Group("/v1")
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))
	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1), h.GetAllUsers)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetUserByID)
	protected.Put("/reset/:id", middleware.RoleMiddleware(authSvc, 1), h.ResetPassword)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateUser)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteUser)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateUsers)

}
