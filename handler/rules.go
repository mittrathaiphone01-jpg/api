package handler

import (
	"rrmobile/service"

	"github.com/gofiber/fiber/v2"
)

type RulesRequestHandler interface {
	GetAllRules(c *fiber.Ctx) error
	GetRuleByID(c *fiber.Ctx) error
	CreateRule(c *fiber.Ctx) error
	EditRule(c *fiber.Ctx) error
	DeleteRule(c *fiber.Ctx) error
}
type rulesHandler struct {
	rulesService service.RulesService
}

func NewRulesHandler(rulesService service.RulesService) *rulesHandler {
	return &rulesHandler{rulesService: rulesService}
}

func (rh *rulesHandler) GetAllRules(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	rules, err := rh.rulesService.GetAllRules(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลกฎได้",
		})
	}

	totalCount, err := rh.rulesService.CountRules()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถนับจำนวนกฎได้",
		})
	}

	currentPage := offset/limit + 1
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	itemsLeft := totalCount - int64(offset+limit)

	response := service.PaginatedRulesResponse{
		Rules:       rules,
		CurrentPage: currentPage,
		TotalPages:  int(totalPages),
		ItemsLeft:   int(itemsLeft),
	}

	return c.JSON(response)
}

func (rh *rulesHandler) GetRuleByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID ไม่ถูกต้อง",
		})
	}

	rule, err := rh.rulesService.GetRuleByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลกฎได้",
		})
	}
	if rule == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบกฎที่ระบุ",
		})
	}

	return c.JSON(rule)
}
func (rh *rulesHandler) CreateRule(c *fiber.Ctx) error {
	var request service.NewRuleRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	rule, err := rh.rulesService.CreateRule(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างกฎได้",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(rule)
}

func (rh *rulesHandler) EditRule(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID ไม่ถูกต้อง",
		})
	}

	var request service.UpdateRuleRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	rule, err := rh.rulesService.EditRule(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถแก้ไขกฎได้",
		})
	}

	return c.JSON(rule)
}
func (rh *rulesHandler) DeleteRule(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID ไม่ถูกต้อง",
		})
	}

	err = rh.rulesService.DeleteRule(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถลบกฎได้",
		})
	}

	return c.JSON(fiber.Map{"message": "กฎถูกลบเรียบร้อยแล้ว"})
}
