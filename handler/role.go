package handler

import (
	"errors"
	"rrmobile/config"
	"rrmobile/service"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RolesRequestHandler interface {
	GetAllRoles(c *fiber.Ctx) error
	CreateRoles(c *fiber.Ctx) error
	UpdateRole(c *fiber.Ctx) error
	DeleteRole(c *fiber.Ctx) error
	GetRoleByID(c *fiber.Ctx) error
}
type rolesHandler struct {
	rolesService service.RolesService
}

func NewRolesHandler(rolesService service.RolesService) *rolesHandler {
	return &rolesHandler{rolesService: rolesService}
}

func (uh *rolesHandler) GetAllRoles(c *fiber.Ctx) error {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	nameStr := c.Query("name") // Example: jack,frank,max
	var name []string
	if nameStr != "" {
		name = strings.Split(nameStr, ",")
	}

	offset := (page - 1) * limit

	// Get total count based on filter
	totalUsers, err := uh.rolesService.CountRoles(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	totalPages := (totalUsers + limit - 1) / limit

	if page > totalPages && totalPages != 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	users, err := uh.rolesService.GetAllRoles(name, limit, offset)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "ไม่พบข้อมูล",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาด",
		})
	}

	nextPage := 0
	prevPage := 0
	if users == nil {
		users = []service.RolesResponseAll{}
	}

	if page < totalPages {
		nextPage = page + 1
	}
	if page > 1 {
		prevPage = page - 1
	}

	response := fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  totalPages,
			"total_users":  totalUsers,
			"next_page":    nextPage,
			"prev_page":    prevPage,
		},
	}

	return c.JSON(response)
}

func (uh *rolesHandler) GetRoleByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}
	role, err := uh.rolesService.GetRoleById(uint(id)) // Call with uint
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}
	return c.JSON(role)
}

func (uh *rolesHandler) CreateRoles(c *fiber.Ctx) error {
	var role service.NewRolesRequest
	if err := c.BodyParser(&role); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	createdRole, err := uh.rolesService.CreateRole(role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_dupli,
		})
		// ✅ กรณี Internal Error อื่น ๆ
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_Intnal, // เกิดข้อผิดพลาดภายในระบบ
		})
	}

	return c.JSON(fiber.Map{
		"data": createdRole,
	})
}

func (uh *rolesHandler) UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	var role service.UpdateRolesRequest // Ensure this struct has the required fields
	if err := c.BodyParser(&role); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ส่งข้อมูลไม่ครบ",
		})
	}

	// Call the service to update the role
	updatedRole, err := uh.rolesService.EditRole(uint(id), role.Role_Name) // Pass the Role_Name directly
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "แก้ไขไม่สำเร็จ",
		})
	}

	// Return the updated role data for confirmation
	return c.JSON(fiber.Map{
		"data": updatedRole,
	})
}

func (uh *rolesHandler) DeleteRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	err = uh.rolesService.DeleteRole(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ลบข้อมูลไม่สำเร็จ",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ลบข้อมูลสำเร็จ",
	})
}
