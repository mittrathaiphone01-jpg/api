package path

import (
	"rrmobile/handler"
	"rrmobile/middleware"
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

func BillPath(app *fiber.App, h handler.BillRequestHandler, authSvc service.AuthService, usersSvc service.UsersService) {

	api := app.Group("/bill")
	v1 := api.Group("/v1")
	v1.Get("/all/unpaid", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllBillsUnpay)
	// protected := v1.Use(middleware.JWTMiddleware(authSvc, usersSvc))
	v1.Post("renew/:id",middleware.RoleMiddleware(authSvc, 1, 2), h.RenewInterest)

	v1.Get("/all", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllBills)
	v1.Get("/all/in", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllInstallmentBills)

	// protected.Get("/all/unpaid", middleware.RoleMiddleware(authSvc, 1), h.GetAllBillsUnpay)
	v1.Get("/all/in/unpaid", middleware.RoleMiddleware(authSvc, 1, 2), h.GetAllInstallmentBillsUnpay)

	v1.Get("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetBillByID)
	v1.Put("/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.UpdateBill)
	v1.Put("/:id/in", middleware.RoleMiddleware(authSvc, 1, 2), h.UpdateBill_Installments)

	v1.Get("/detail/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetBillDetailByID)
	v1.Get("all/detail/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetDetailBillByBillID)
	v1.Get("all/detail/in/:id", middleware.RoleMiddleware(authSvc, 1, 2), h.GetInstallmentBillByIdUnpaid)

	v1.Get("/detail/:id/in", middleware.RoleMiddleware(authSvc, 1, 2), h.GetInstallmentBillDetailByID)

	v1.Post("/create", middleware.RoleMiddleware(authSvc, 1, 2), h.CreateBill)
	// v1.Post("/refresh", h.Refresh)
	v1.Post("/createst", h.CreateBill)

	v1.Post("/create/in", middleware.RoleMiddleware(authSvc, 1, 2), h.CreateBillInsallments)
	v1.Get("/:id/in", middleware.RoleMiddleware(authSvc, 1, 2), h.GetInstallmentBillByID)

	v1.Post("/extra", h.AddExtraPayment, middleware.RoleMiddleware(authSvc, 1, 2))
	v1.Post("/extra/in", h.AddInstallmentExtraPayment, middleware.RoleMiddleware(authSvc, 1, 2))
	private := v1.Group("/", middleware.RequireBillAuth())
	private.Get("/unpaid/today", h.GetDueTodayBillsHandler)
	private.Get("/unpaid/today/in", h.GetDueTodayInstallmentBillsHandler)
	private.Post("/all/unpaid/bill", h.GetUnpaidBillByIdHandler)
	private.Post("/all/unpaid/bill/in", h.GetUnpaidInstallmentBillByIdHandler)

	private.Post("/paid/bill", h.GetpaidBillByIdHandler)
	private.Post("/paid/bill/in", h.GetpaidInstallBillByIdHandler)

	private.Post("/pay", h.PayInstallment)
	private.Post("/pay/in", h.PayInstallmentBill)

}
