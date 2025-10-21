package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func ProductPath(app *fiber.App, h handler.ProductRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/product")
	v1 := api.Group("/v1")
	// private := v1.Group("/", middleware.RequireBillAuth())
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))

	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllProducts)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetProductByID)
	protected.Get("/:id/detail", middleware.RoleMiddleware(authSvc, 1, 2), h.GetProductByIDDetail)

	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1, 2), h.AddProduct)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1), h.UpdateProduct)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1), h.DeleteProduct)
	// v1.Get("/category/:categoryId", h.GetProductsByCategory)
	// v1.Get("/name/:name", h.GetProductsByName)
	// v1.Get("/description/:description", h.GetProductsByDescription)
}

func ProductCategoryPath(app *fiber.App, h handler.ProductCategoryHandler, authSvc service.AuthService, usersSvc service.UsersService) {
	api := app.Group("/product-category")
	v1 := api.Group("/v1")
	protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))
	// v1.Use(middleware.MiddlewareJWT())
	// v1.Use(middleware.AuthMiddleware(1))
	protected.Get("/all", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllCategories)
	protected.Get("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetCategoryByID)
	protected.Post("/create", middleware.RoleMiddleware(authSvc, 1, 2), h.CreateCategory)
	protected.Put("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.UpdateCategory)
	protected.Delete("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.DeleteCategory)
	// v1.Get("/name/:name", h.GetCategoriesByName)
}
