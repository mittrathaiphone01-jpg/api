package service

import "time"

type UserResponse struct {
	Id        uint   `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	RoleID    int    `json:"role_id"`
	Is_active bool   `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserResponseAll struct {
	Id        uint   `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	RoleID    int    `json:"role_id"`
	Is_active bool   `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type NewUserRequest struct {
	Username  string `json:"username"  validate:"required"`
	Password  string `json:"Password"  validate:"required"`
	FullName  string `json:"full_name" validate:"required"`
	RoleID    uint   `json:"role_id"`
	Is_active bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Password  string    `json:"Password"  validate:"required"`
	UpdatedAt time.Time `json:"updated_at"`
	Is_active *bool     `json:"is_active"` // ✅ ใช้ pointer

}

type UpdateUserPasswordRequest struct {
	Password string `json:"Password"`
}
type UsersService interface {
	GetAllUsers(usernames []string, date string, limit, offset int) ([]UserResponseAll, error)
	// GetAllUsers(limit int, offset int) ([]UserResponseAll, error)
	// CountUsers() (int, error)
	CountUsers(usernames []string, date string) (int, error)
	GetUserById(id uint) (*UserResponse, error)
	CreateUser(NewUserRequest) (*UserResponse, error)
	EditUser(id uint, newUsername string, newFullName string, newPassword string, newActive *bool) (*UserResponse, error)
	DeleteUser(id uint) error
	ResetPasswordUser(id uint, newPassword string) (*UserResponse, error)
}
