package handler

import (
	"rrmobile/service"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type MemberRequestHandler interface {
	GetMembers(c *fiber.Ctx) error
	GetMemberById(c *fiber.Ctx) error
	CreateMember(c *fiber.Ctx) error
	UpdateMember(c *fiber.Ctx) error
	DeleteInstallment(c *fiber.Ctx) error

	GetMemberByUserId(c *fiber.Ctx) error
	LinkUserByTel(c *fiber.Ctx) error
}

type memberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(memberService service.MemberService) *memberHandler {
	return &memberHandler{memberService: memberService}
}

func (h *memberHandler) GetMembers(c *fiber.Ctx) error {
	full_name := strings.Split(c.Query("full_name", ""), ",")
	user_id := strings.Split(c.Query("user_id", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", ""))

	resp, err := h.memberService.GetAllMembers(
		full_name, user_id, page, limit,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
func (ih *memberHandler) GetMemberById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	installment, err := ih.memberService.GetMemberById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลที่ต้องการ"})
	}
	return c.JSON(installment)
}
func (ih *memberHandler) CreateMember(c *fiber.Ctx) error {
	var request service.NewMemberRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	member, err := ih.memberService.CreateMember(request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างข้อมูลได้"})
	}
	return c.Status(fiber.StatusCreated).JSON(member)
}
func (ih *memberHandler) UpdateMember(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	var request service.UpdateMemberRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	member, err := ih.memberService.EditMember(uint(id), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทข้อมูลได้"})
	}
	return c.JSON(member)
}

func (ih *memberHandler) DeleteInstallment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	if err := ih.memberService.DeleteMember(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลบข้อมูลได้"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (ih *memberHandler) GetMemberByUserId(c *fiber.Ctx) error {
	type Request struct {
		UserID string `json:"user_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id ห้ามว่าง"})
	}

	member, err := ih.memberService.GetMemberByUserId(req.UserID)
	if err != nil {
		// ถ้าไม่พบข้อมูลให้ส่ง false
		return c.JSON(fiber.Map{"data": false})
	}

	// ถ้าพบข้อมูลให้ส่ง true
	return c.JSON(fiber.Map{"data": true, "member": member})
}

func (h *memberHandler) LinkUserByTel(c *fiber.Ctx) error {
	type Request struct {
		Tel    *string `json:"tel"`
		UserId *string `json:"user_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// ตรวจสอบว่ามีอย่างน้อย 1 field ถูกส่งมา
	if req.Tel == nil && req.UserId == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one of 'tel' or 'user_id' is required",
		})
	}

	// เรียก service
	result, err := h.memberService.UpdateMemberByTel(service.UpdateMemberRequest1{
		Tel:    req.Tel,
		UserId: req.UserId,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Member updated",
		"data":    result,
	})
}
