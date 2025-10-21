package service

import (
	"fmt"
	"rrmobile/respository"
)

type rolesService struct {
	rolesRepository respository.RoleRepository
}

func NewRoleService(rolesRepository respository.RoleRepository) RolesService {
	return &rolesService{rolesRepository: rolesRepository}
}
func (s *rolesService) GetAllRoles(name []string, limit, offset int) ([]RolesResponseAll, error) {
	roles, err := s.rolesRepository.GetAllRoles(name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve roles")
	}

	var rolesResponse []RolesResponseAll
	for _, role := range roles {
		rolesResponse = append(rolesResponse, RolesResponseAll{
			Id:        role.Id,
			Role_Name: role.Role_Name,
		})
	}
	return rolesResponse, nil
}
func (s *rolesService) CountRoles(name []string) (int, error) {
	var count int64
	err := s.rolesRepository.CountRoles(name, &count)
	if err != nil {
		return 0, fmt.Errorf("Failed to count roles")
	}
	return int(count), nil
}
func (s *rolesService) GetRoleById(id uint) (*RolesResponse, error) {
	role, err := s.rolesRepository.GetRoleByID(id)
	if err != nil {
		return nil, fmt.Errorf("Role with ID not found", id)
	}
	response := RolesResponse{
		Id:        role.Id,
		Role_Name: role.Role_Name,
	}
	return &response, nil
}

func (s *rolesService) CreateRole(request NewRolesRequest) (*RolesResponse, error) {
	// ตรวจสอบว่า Role Name ถูกส่งมาหรือไม่
	if request.Role_Name == "" {
		return nil, fmt.Errorf("Please provide a role name")
	}

	// ตรวจสอบจำนวน Role ปัจจุบัน

	// ตรวจสอบว่ามี Role ซ้ำในฐานข้อมูลแล้วหรือไม่ และเพิ่ม Role ใหม่
	role := respository.Roles{
		Role_Name: request.Role_Name,
	}
	newRole, err := s.rolesRepository.AddRole(role)
	if err != nil {
		return nil, fmt.Errorf("Failed to create role")
	}

	// สร้าง response
	response := RolesResponse{
		Id:        newRole.Id,
		Role_Name: newRole.Role_Name,
	}
	return &response, nil
}

func (s *rolesService) EditRole(id uint, newRoleName string) (*RolesResponse, error) {
	role, err := s.rolesRepository.GetRoleByID(id)
	if err != nil {
		return nil, fmt.Errorf("Role with ID not found", id)
	}

	role.Role_Name = newRoleName
	updatedRole, err := s.rolesRepository.UpdateRole(id, *role) // Pass the updated role
	if err != nil {
		return nil, fmt.Errorf("Failed to update role")
	}

	response := RolesResponse{
		Id:        updatedRole.Id,
		Role_Name: updatedRole.Role_Name,
	}
	return &response, nil
}
func (s *rolesService) DeleteRole(id uint) error {
	return s.rolesRepository.DeleteRole(id)
}
