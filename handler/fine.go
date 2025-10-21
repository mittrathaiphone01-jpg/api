package handler

import (
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

type FineRequestHandler interface {
	GetFines(c *fiber.Ctx) error
	GetFineById(c *fiber.Ctx) error
	CreateFine(c *fiber.Ctx) error
	UpdateFine(c *fiber.Ctx) error
	DeleteFine(c *fiber.Ctx) error
}
type fineHandler struct {
	fineService service.FineService
}

func NewFineHandler(fineService service.FineService) *fineHandler {
	return &fineHandler{fineService: fineService}
}

func (ih *fineHandler) GetFines(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	fines, err := ih.fineService.GetFines(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลได้"})
	}
	return c.JSON(fines)
}

func (ih *fineHandler) GetFineById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	fine, err := ih.fineService.GetFineById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}
	return c.JSON(fine)
}
func (ih *fineHandler) CreateFine(c *fiber.Ctx) error {
	var request service.NewFineRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	installment, err := ih.fineService.CreateFine(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างข้อมูลได้"})
	}
	return c.Status(fiber.StatusCreated).JSON(installment)
}
func (ih *fineHandler) UpdateFine(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	var request service.UpdateFineRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	installment, err := ih.fineService.UpdateFine(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทข้อมูลได้"})
	}
	return c.JSON(installment)
}

func (ih *fineHandler) DeleteFine(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	if err := ih.fineService.DeleteFine(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลบข้อมูลได้"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
