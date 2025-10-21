package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func MemberPath(app *fiber.App, h handler.MemberRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/members")
	v1 := api.Group("/v1")
	v1.Get("/all", middleware.RoleMiddleware(authSvc, 1, 2), h.GetMembers)
	v1.Get("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetMemberById)
	v1.Post("/create", middleware.RoleMiddleware(authSvc, 1, 2), h.CreateMember)
	v1.Put("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.UpdateMember)
	v1.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteInstallment)
	private := v1.Group("/", middleware.RequireBillAuth())
	private.Post("/checking", h.GetMemberByUserId)
	private.Post("/link", h.LinkUserByTel)
	// protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

}
