package handler

import (
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

type InstallmentRequestHandler interface {
	GetInstallments(c *fiber.Ctx) error
	GetInstallmentById(c *fiber.Ctx) error
	CreateInstallment(c *fiber.Ctx) error
	UpdateInstallment(c *fiber.Ctx) error
	DeleteInstallment(c *fiber.Ctx) error
}

type installmentHandler struct {
	installmentService service.InstallmentService
}

func NewInstallmentHandler(installmentService service.InstallmentService) *installmentHandler {
	return &installmentHandler{installmentService: installmentService}
}

func (ih *installmentHandler) GetInstallments(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	installments, err := ih.installmentService.GetInstallments(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลได้"})
	}
	return c.JSON(installments)
}

func (ih *installmentHandler) GetInstallmentById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	installment, err := ih.installmentService.GetInstallmentById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}
	return c.JSON(installment)
}
func (ih *installmentHandler) CreateInstallment(c *fiber.Ctx) error {
	var request service.NewInstallmentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	installment, err := ih.installmentService.CreateInstallment(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างข้อมูลได้"})
	}
	return c.Status(fiber.StatusCreated).JSON(installment)
}

func (ih *installmentHandler) UpdateInstallment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	var request service.UpdateInstallmentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	installment, err := ih.installmentService.UpdateInstallment(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทข้อมูลได้"})
	}

	return c.JSON(installment)
}

func (ih *installmentHandler) DeleteInstallment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	if err := ih.installmentService.DeleteInstallment(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลบข้อมูลได้"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
