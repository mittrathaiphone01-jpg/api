package path

import (
	"rrmobile/handler"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func AuthPath(app *fiber.App, h handler.AuthRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/auth")
	v1 := api.Group("/v1")
	v1.Post("/login", h.Login)
	v1.Post("/logout", h.Logout)
	v1.Post("/refresh", h.Refresh)

}
