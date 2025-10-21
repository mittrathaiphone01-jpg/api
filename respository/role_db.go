package respository

import (
	"fmt"

	"gorm.io/gorm"
)

type roleRespositoryDB struct {
	db *gorm.DB
}

func NewRoleRepositoryDB(db *gorm.DB) RoleRepository {
	return &roleRespositoryDB{db: db}
}

func (r *roleRespositoryDB) CountRoles(name []string, count *int64) error {
	query := r.db.Model(&Roles{})
	if len(name) > 0 {
		query = query.Where("name IN ?", name)
	}

	return r.db.Model(&Roles{}).Count(count).Error
}
func (r *roleRespositoryDB) GetAllRoles(name []string, limit int, offset int) ([]Roles, error) {

	var roles []Roles
	query := r.db.Model(&Roles{})
	if len(name) > 0 {
		query = query.Where("role_name IN ?", name)
	}

	if limit <= 0 {
		limit = 10 // กำหนดค่าเริ่มต้นถ้า limit น้อยกว่าหรือเท่ากับ 0
	}
	if offset < 0 {
		offset = 0 // กำหนดค่าเริ่มต้นถ้า offset น้อยกว่า 0
	}
	err := r.db.Order("id DESC").Offset(offset).Limit(limit).Find(&roles).Error

	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		// หากไม่พบข้อมูล สามารถคืนค่า error หรือ nil ได้ตามที่คุณต้องการ
		return nil, fmt.Errorf("no roles found")
	}
	return roles, nil
}
func (r *roleRespositoryDB) GetRoleByID(id uint) (*Roles, error) {
	var role Roles
	err := r.db.Where("id =?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRespositoryDB) AddRole(role Roles) (*Roles, error) {
	// เพิ่มข้อมูลในฐานข้อมูลโดย GORM จะทำการ auto-increment ให้
	existingRole := Roles{}

	tx := r.db.Where("role_name = ? ", role.Role_Name).First(&existingRole)
	if tx.RowsAffected > 0 {
		return nil, fmt.Errorf("This role already exists")
	}
	err := r.db.Create(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil // คืนค่า role ที่ถูกบันทึกพร้อมกับ Primary Key ที่เพิ่มอัตโนมัติ
}

func (r *roleRespositoryDB) UpdateRole(id uint, role Roles) (*Roles, error) {
	var oldRole Roles
	err := r.db.Where("id = ?", id).First(&oldRole).Error
	if err != nil {
		return nil, err
	}

	// Update the fields of oldRole with new values
	oldRole.Role_Name = role.Role_Name // Assuming you only want to update the Role_Name

	// Save the updated role
	err = r.db.Save(&oldRole).Error
	if err != nil {
		return nil, err
	}
	return &oldRole, nil
}

func (r *roleRespositoryDB) DeleteRole(id uint) error {
	var role Roles
	err := r.db.Where("id =?", id).Delete(&role).Error
	if err != nil {
		return err
	}
	return nil
}
