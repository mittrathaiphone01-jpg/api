package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func FineCategoryPath(app *fiber.App, h handler.FineCategoryRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {

	api := app.Group("/finecategory")
	v1 := api.Group("/v1")
	// private := v1.Group("/", middleware.RequireBillAuth())
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1), h.GetFineCategories)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1), h.GetFineCategoryById)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1), h.CreateFineCategory)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateFineCategory)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteFineCategory)

}
