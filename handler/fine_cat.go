package handler

import (
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

type FineCategoryRequestHandler interface {
	GetFineCategories(c *fiber.Ctx) error
	GetFineCategoryById(c *fiber.Ctx) error
	CreateFineCategory(c *fiber.Ctx) error
	UpdateFineCategory(c *fiber.Ctx) error
	DeleteFineCategory(c *fiber.Ctx) error
}
type fineCategoryHandler struct {
	fineCategoryService service.FineCategoryService
}

func NewFineCategoryHandler(fineCategoryService service.FineCategoryService) *fineCategoryHandler {
	return &fineCategoryHandler{fineCategoryService: fineCategoryService}
}
func (ih *fineCategoryHandler) GetFineCategories(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	categories, err := ih.fineCategoryService.GetFineCategories(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลได้"})
	}
	return c.JSON(categories)
}
func (ih *fineCategoryHandler) GetFineCategoryById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	category, err := ih.fineCategoryService.GetFineCategoryById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}
	return c.JSON(category)
}
func (ih *fineCategoryHandler) CreateFineCategory(c *fiber.Ctx) error {
	var request service.NewFineCategoryRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	category, err := ih.fineCategoryService.CreateFineCategory(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างข้อมูลได้"})
	}
	return c.JSON(category)
}
func (ih *fineCategoryHandler) UpdateFineCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	var request service.UpdateFineCategoryRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	category, err := ih.fineCategoryService.UpdateFineCategory(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัปเดตข้อมูลได้"})
	}
	return c.JSON(category)
}
func (ih *fineCategoryHandler) DeleteFineCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	if err := ih.fineCategoryService.DeleteFineCategory(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลบข้อมูลได้"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
