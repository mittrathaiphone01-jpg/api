package respository

import "time"

type User struct {
	Id        uint      `db:"id"`
	Username  string    `db:"username"`
	FullName  string    `db:"full_name"`
	Password  string    `db:"password"`
	RoleID    uint      `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Is_active *bool     `db:"is_active"` // ✅ ใช้ pointer

}

type UserRepository interface {
	GetAllUsers(usernames []string, date string, limit, offset int) ([]User, error)
	GetUserByID(id uint) (*User, error)
	CountUsers(usernames []string, date string, count *int64) error
	AddUser(user User) (*User, error)
	UpdateUser(id uint, user User) (*User, error)
	DeleteUser(id uint) error
	UpdatePassword(id uint, newPassword string) (*User, error)
}
