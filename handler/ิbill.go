package handler

import (
	"log"
	"rrmobile/config"
	"rrmobile/service"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BillRequestHandler interface {
	CreateBill(c *fiber.Ctx) error
	PayInstallment(c *fiber.Ctx) error
	AddExtraPayment(c *fiber.Ctx) error
	GetAllBills(c *fiber.Ctx) error
	GetBillByID(c *fiber.Ctx) error
	GetBillDetailByID(c *fiber.Ctx) error
	CreateBillInsallments(c *fiber.Ctx) error
	GetInstallmentBillByID(c *fiber.Ctx) error
	PayInstallmentBill(c *fiber.Ctx) error
	AddInstallmentExtraPayment(c *fiber.Ctx) error
	GetAllInstallmentBills(c *fiber.Ctx) error
	UpdateBill(c *fiber.Ctx) error
	UpdateBill_Installments(c *fiber.Ctx) error
	GetInstallmentBillDetailByID(c *fiber.Ctx) error
	GetAllBillsUnpay(c *fiber.Ctx) error
	GetAllInstallmentBillsUnpay(c *fiber.Ctx) error
	GetDetailBillByBillID(c *fiber.Ctx) error
	GetInstallmentBillByIdUnpaid(c *fiber.Ctx) error

	GetDueTodayBillsHandler(c *fiber.Ctx) error
	GetDueTodayInstallmentBillsHandler(c *fiber.Ctx) error

	GetUnpaidBillByIdHandler(c *fiber.Ctx) error
	GetUnpaidInstallmentBillByIdHandler(c *fiber.Ctx) error

	GetpaidBillByIdHandler(c *fiber.Ctx) error
	GetpaidInstallBillByIdHandler(c *fiber.Ctx) error

	RenewInterest(c *fiber.Ctx) error
}
type billHandler struct {
	billService service.BillService
}

func NewBillHandler(billService service.BillService) *billHandler {
	return &billHandler{billService: billService}
}

func (ih *billHandler) CreateBill(c *fiber.Ctx) error {
	var request service.NewBillHeader
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	category, err := ih.billService.CreateBill(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างข้อมูลได้"})
	}
	return c.JSON(category)
}

func (h *billHandler) PayInstallment(c *fiber.Ctx) error {
	type Request struct {
		BillID       uint    `json:"bill_id"`
		BillDetailID uint    `json:"bill_detail_id"`
		Amount       float64 `json:"amount"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.BillID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bill_id and amount are required",
		})
	}

	results, err := h.billService.PayInstallment(req.BillID, req.BillDetailID, req.Amount)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ส่งกลับผลลัพธ์ของแต่ละงวดพร้อม Case และเครดิตที่เหลือ
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "installment processed successfully",
		"results": results,
		"bill_id": req.BillID,
		"amount":  req.Amount,
	})
}

func (h *billHandler) AddExtraPayment(c *fiber.Ctx) error {
	var req service.UpdateAddExtraRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// เรียก service
	if err := h.billService.AddExtraPayment(req.BillID, req.InstallmentID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":        "extra payment added successfully",
		"bill_id":        req.BillID,
		"installment_id": req.InstallmentID,
		"extra_amount":   req.Paid_Amount,
	})
}

func (h *billHandler) GetAllBills(c *fiber.Ctx) error {
	skus := strings.Split(c.Query("search", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "0"))    // default 0 = all
	sortOrder, _ := strconv.Atoi(c.Query("sort", "0")) // 0 = default DESC, 1 = DESC, 2 = ASC

	// Optional filters
	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}

	resp, err := h.billService.GetAllBill(
		skus, dateFrom, dateTo, page, limit, sortOrder, // ส่ง sortOrder ไป service
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
func (h *billHandler) GetAllBillsUnpay(c *fiber.Ctx) error {
	skus := strings.Split(c.Query("search", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "0"))    // default 0 = all
	sortOrder, _ := strconv.Atoi(c.Query("sort", "0")) // 0 = default DESC, 1 = DESC, 2 = ASC

	// Optional filters
	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}
	np := c.Query("np", "")
	npParts := strings.Split(np, ",")

	resp, err := h.billService.GetAllBillUnpay(
		skus, dateFrom, dateTo, page, limit, sortOrder, npParts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
func (h *billHandler) GetBillByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetBillById(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}
func (h *billHandler) GetBillDetailByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetBillDetailById(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}

func (ih *billHandler) CreateBillInsallments(c *fiber.Ctx) error {
	var request service.NewInstallmentBillHeader

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}


	resp, err := ih.billService.CreateInstallmentBill(request, uint(request.InstallmentId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(), // 👈 โชว์ error จริงไว้ก่อน (ตอน dev)
		})
	}

	return c.JSON(resp)
}

func (h *billHandler) GetInstallmentBillByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetInstallmentBillById(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}
func (h *billHandler) PayInstallmentBill(c *fiber.Ctx) error {
	type Request struct {
		BillID       uint `json:"bill_id"`
		BillDetailID uint `json:"bill_detail_id"`

		Amount float64 `json:"amount"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.BillID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bill_id and amount are required",
		})
	}

	// เรียก service
	results, err := h.billService.PayPurchaseInstallment(req.BillID, req.BillDetailID, req.Amount)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ส่ง JSON Response กลับ
	return c.JSON(fiber.Map{
		"data": results,
	})
}

func (h *billHandler) AddInstallmentExtraPayment(c *fiber.Ctx) error {
	var req service.UpdateAddExtraRequest_Installment
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// เรียก service
	if err := h.billService.AddInstallmentExtraPayment(req.BillID, req.InstallmentID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":        "extra payment added successfully",
		"bill_id":        req.BillID,
		"installment_id": req.InstallmentID,
		"extra_amount":   req.Paid_Amount,
	})
}

func (h *billHandler) GetAllInstallmentBills(c *fiber.Ctx) error {
	skus := strings.Split(c.Query("search", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "0"))    // default 0 = all
	sortOrder, _ := strconv.Atoi(c.Query("sort", "0")) // 0 = default DESC, 1 = DESC, 2 = ASC

	// Optional filters
	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}

	resp, err := h.billService.GetAllInstallmentBill(
		skus, dateFrom, dateTo, page, limit, sortOrder, // ส่ง sortOrder ไป service
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
func (h *billHandler) GetAllInstallmentBillsUnpay(c *fiber.Ctx) error {
	skus := strings.Split(c.Query("search", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "0"))    // default 0 = all
	sortOrder, _ := strconv.Atoi(c.Query("sort", "0")) // 0 = default DESC, 1 = DESC, 2 = ASC

	// Optional filters
	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}
	// nameOrPhone := c.Query("np", "") // รับค่า `name_or_phone` จาก query string
	np := c.Query("np", "")
	npParts := strings.Split(np, ",")

	resp, err := h.billService.GetAllInstallmentBillUnpay(
		skus, dateFrom, dateTo, page, limit, sortOrder, npParts,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}

func (h *billHandler) UpdateBill(c *fiber.Ctx) error {
	// ดึง id จาก param
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bill id",
		})
	}

	// bind body เข้ากับ request struct
	var request service.Update_Installment
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// เรียก service
	updatedBill, err := h.billService.UpdateBill(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// success response
	return c.Status(fiber.StatusOK).JSON(updatedBill)
}

func (h *billHandler) UpdateBill_Installments(c *fiber.Ctx) error {
	// ดึง id จาก param
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bill id",
		})
	}

	// bind body เข้ากับ request struct
	var request service.Update_Installment
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// เรียก service
	updatedBill, err := h.billService.UpdateBill_Installment(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// success response
	return c.Status(fiber.StatusOK).JSON(updatedBill)
}

func (h *billHandler) GetInstallmentBillDetailByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetInstallmentBillDetailById(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}

func (h *billHandler) GetDetailBillByBillID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetBillDetailsByIdUnpaid(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}
func (h *billHandler) GetInstallmentBillByIdUnpaid(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := h.billService.GetInstallmentBillByIdUnpaid(uint(id))
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal,
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
	// return c.JSON
}

func (h *billHandler) GetDueTodayBillsHandler(c *fiber.Ctx) error {
	// รับ query param "sort" เท่านั้น
	sortData := c.Query("sort", "asc")

	// เรียก service
	result, err := h.billService.GetDueTodayBillsWithInstallments(sortData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetDueTodayInstallmentBillsHandler(c *fiber.Ctx) error
func (h *billHandler) GetDueTodayInstallmentBillsHandler(c *fiber.Ctx) error {
	// รับ query param "sort" เท่านั้น
	sortData := c.Query("sort", "asc")

	// เรียก service
	result, err := h.billService.GetDueTodayInstallmentBillsWithInstallments(sortData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

type GetUnpaidBillRequest struct {
	UserId string `json:"user_id"`
}
type GetUnpaidInstallmentBillRequest struct {
	UserId string `json:"user_id"`
}

func (h *billHandler) GetUnpaidBillByIdHandler(c *fiber.Ctx) error {
	var req GetUnpaidBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ไม่สามารถอ่านข้อมูลได้"})
	}

	if req.UserId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุ user_id"})
	}

	installment, err := h.billService.GetUnpaidBillById(req.UserId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}

	return c.JSON(installment)
}

func (h *billHandler) GetUnpaidInstallmentBillByIdHandler(c *fiber.Ctx) error {
	var req GetUnpaidInstallmentBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ไม่สามารถอ่านข้อมูลได้"})
	}

	// id, err := c.ParamsInt("id")
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	// }
	installment, err := h.billService.GetUnpaidInstallmentBillById(req.UserId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}
	return c.JSON(installment)
}

type GetpaidBillRequest struct {
	BillID   uint `json:"bill_id"`
	DetailID uint `json:"bill_detail_id"`
}

func (h *billHandler) GetpaidBillByIdHandler(c *fiber.Ctx) error {
	var req GetpaidBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ไม่สามารถอ่านข้อมูลได้"})
	}
	if req.BillID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุ ข้อมูล"})
	}

	if req.DetailID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุ ข้อมูล"})
	}

	installment, err := h.billService.GetpaidBillById(req.BillID, req.DetailID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}

	return c.JSON(installment)
}

func (h *billHandler) GetpaidInstallBillByIdHandler(c *fiber.Ctx) error {

	var req GetpaidBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่สามารถอ่านข้อมูลได้",
		})
	}

	if req.BillID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุข้อมูล BillID"})
	}

	if req.DetailID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุข้อมูล DetailID"})
	}

	installment, err := h.billService.GetpaidInstallmentBillById(req.BillID, req.DetailID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลที่ต้องการ",
		})
	}

	return c.JSON(installment)
}

type RenewInterestRequest struct {
	PayAmount float64 `json:"pay_amount"`
	payNext string  `json:"pay_date"`
	PayDate   string  `json:"pay_date"`
}

// func (h *billHandler) RenewInterest(c *fiber.Ctx) error {
// 	billIDParam := c.Params("id")
// 	billID, err := strconv.ParseUint(billIDParam, 10, 32)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "billID ไม่ถูกต้อง"})
// 	}

// 	var req RenewInterestRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ครบหรือไม่ถูกต้อง"})
// 	}

// 	var payDate time.Time
// 	if req.PayDate == "" {
// 		payDate = time.Now()
// 	} else {
// 		payDate, err = time.Parse("2006-01-02", req.PayDate)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบวันที่ไม่ถูกต้อง"})
// 		}
// 	}

	
// 	log.Print("patDate", payDate)
// 	bill, err := h.billService.RenewInterest(uint(billID), req.PayAmount, payDate)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.JSON(bill)

// 	// return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 	// 	"message": "ต่อดอกสำเร็จ",
// 	// 	"data":    bill,
// 	// })
// }
func (h *billHandler) RenewInterest(c *fiber.Ctx) error {
	billIDParam := c.Params("id")
	billID, err := strconv.ParseUint(billIDParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "billID ไม่ถูกต้อง"})
	}

	var req RenewInterestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ครบหรือไม่ถูกต้อง"})
	}

	var payDate time.Time
	if req.PayDate == "" {
		payDate = time.Now()
	} else {
		payDate, err = time.Parse("2006-01-02", req.PayDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบวันที่ไม่ถูกต้อง"})
		}
	}

	log.Print("payDate", payDate)
	_, err = h.billService.RenewInterest(uint(billID), req.PayAmount, payDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ต่อดอกสำเร็จ",
	})
}
