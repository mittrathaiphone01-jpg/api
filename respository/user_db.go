package respository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type userRespositoryDB struct {
	db *gorm.DB
}

// NewUserRepositoryDB สร้างและคืนค่า UserRepository ใหม่
func NewUserRepositoryDB(db *gorm.DB) UserRepository {
	return &userRespositoryDB{db: db}
}

func (r *userRespositoryDB) CountUsers(usernames []string, date string, count *int64) error {
	query := r.db.Model(&User{})
	query = query.Where("role_id != ?", 1)

	if len(usernames) > 0 {
		orConditions := ""
		var orValues []interface{}

		for i, kw := range usernames {
			if i > 0 {
				orConditions += " OR "
			}
			orConditions += "(username ILIKE ? OR full_name ILIKE ?)"
			pattern := "%" + kw + "%"
			orValues = append(orValues, pattern, pattern)
		}
		query = query.Where(orConditions, orValues...)
	}

	if date != "" {
		query = query.Where("DATE(created_at) = ?", date)
	}

	return query.Count(count).Error
}

func (r *userRespositoryDB) GetAllUsers(usernames []string, date string, limit, offset int) ([]User, error) {
	var users []User
	query := r.db.Model(&User{})
	query = query.Where("role_id != ?", 1)
	if len(usernames) > 0 {
		orConditions := ""
		var orValues []interface{}

		for i, kw := range usernames {
			if i > 0 {
				orConditions += " OR "
			}
			orConditions += "(username ILIKE ? OR full_name ILIKE ?)"
			pattern := "%" + kw + "%"
			orValues = append(orValues, pattern, pattern)
		}

		query = query.Where(orConditions, orValues...)
	}
	if date != "" {
		// Assuming your date is stored as DATE (not DATETIME)
		query = query.Where("DATE(created_at) = ?", date)
	}

	// ไม่สนใจ date เลยข้ามไป

	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users, nil
}


func (r *userRespositoryDB) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.Where("id =?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRespositoryDB) AddUser(user User) (*User, error) {
	existingUser := User{}

	// ✅ แก้ WHERE ให้ถูกต้อง
	tx := r.db.Where("username = ? AND full_name = ?", user.Username, user.FullName).First(&existingUser)
	if tx.RowsAffected > 0 {
		return nil, fmt.Errorf("This data already exists")
	}

	// ✅ เพิ่ม transaction (optional)
	err := r.db.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRespositoryDB) UpdateUser(id uint, user User) (*User, error) {
	var oldUser User
	err := r.db.Where("id =?", id).First(&oldUser).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Model(&oldUser).Updates(user).Error
	if err != nil {
		if strings.Contains(err.Error(), "username") {
			return nil, fmt.Errorf("username already exists")
		}
		if strings.Contains(err.Error(), "full_name") {
			return nil, fmt.Errorf("full_name already exists")
		}
		return nil, err
	}

	return &oldUser, nil
}

func (r *userRespositoryDB) DeleteUser(id uint) error {
	var user User
	err := r.db.Where("id =?", id).Delete(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userRespositoryDB) UpdatePassword(id uint, newPassword string) (*User, error) {
	var user User
	// Find the user by ID
	err := r.db.Where("id =?", id).First(&user).Error
	if err != nil {
		return nil, err
	}

	// Update the password field
	user.Password = newPassword
	err = r.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
