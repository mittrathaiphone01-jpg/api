package handler

import (
	"rrmobile/config"
	"rrmobile/service"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type UsersRequestHandler interface {
	GetAllUsers(c *fiber.Ctx) error
	GetUserByID(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	CreateUsers(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
}
type usersHandler struct {
	usersService service.UsersService
}

func NewUsersHandler(usersService service.UsersService) *usersHandler {
	return &usersHandler{usersService: usersService}
}

func (uh *usersHandler) GetAllUsers(c *fiber.Ctx) error {
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

	usernamesStr := c.Query("search")
	var usernames []string
	if usernamesStr != "" {
		usernames = strings.Split(usernamesStr, ",")
	}

	date := c.Query("date") // Example: 2025-06-26

	offset := (page - 1) * limit

	// Get total count based on filter
	totalUsers, err := uh.usersService.CountUsers(usernames, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	totalPages := (totalUsers + limit - 1) / limit

	if page > totalPages && totalPages != 0 {
		// ถ้าหน้าเกินจริง ให้คืน data=[] และ pagination ที่ถูกต้อง
		return c.JSON(fiber.Map{
			"data": []service.UserResponseAll{},
			"pagination": fiber.Map{
				"current_page": page,
				"total_pages":  totalPages,
				"total_users":  totalUsers,
				"next_page":    0,
				"prev_page":    0,
			},
		})
	}

	users, err := uh.usersService.GetAllUsers(usernames, date, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาด",
		})
	}

	// ✅ จุดสำคัญ: ป้องกันไม่ให้ data เป็น null
	if users == nil {
		users = []service.UserResponseAll{}
	}

	nextPage := 0
	prevPage := 0

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

func (uh *usersHandler) GetUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := uh.usersService.GetUserById(uint(id))
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

func (uh *usersHandler) CreateUsers(c *fiber.Ctx) error {
	var user service.NewUserRequest
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_BodyReq,
		})

	}
	createdUser, err := uh.usersService.CreateUser(user)
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
		"data": createdUser,
	})
}

func (uh *usersHandler) UpdateUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_Invail,
		})
	}

	var user service.UpdateUserRequest // Ensure this struct has the required fields
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_BodyReq,
		})
	}

	// Call the service to update the role
	updatedRole, err := uh.usersService.EditUser(uint(id), user.Username, user.FullName, user.Password, user.Is_active)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_update,
		})
	}

	// Return the updated role data for confirmation
	return c.JSON(fiber.Map{
		"data": updatedRole,
	})
}

func (uh *usersHandler) DeleteUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_Invail,
		})
	}

	err = uh.usersService.DeleteUser(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_del,
		})
	}

	return c.SendStatus(fiber.StatusNoContent) // 204 No Content status code
}

func (uh *usersHandler) ResetPassword(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_Invail,
		})
	}

	var user service.UpdateUserPasswordRequest // Ensure this struct has the required fields
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": config.Error_BodyReq,
		})
	}

	// Call the service to update the role
	updatedRole, err := uh.usersService.ResetPasswordUser(uint(id), user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": config.Error_update,
		})
	}

	// Return the updated role data for confirmation
	return c.JSON(fiber.Map{
		"data": updatedRole,
	})
}
