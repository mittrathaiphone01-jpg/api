package service

type RolesResponse struct {
	Id        uint   `json:"id"`
	Role_Name string `json:"role_name"`
}

type RolesResponseAll struct {
	Id          uint   `json:"id"`
	Role_Name   string `json:"role_name"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	ItemsLeft   int    `json:"items_left"`
}
type PaginatedRolesResponse struct {
	Roles       []RolesResponse `json:"roles"`
	CurrentPage int             `json:"current_page"`
	TotalPages  int             `json:"total_pages"`
	ItemsLeft   int             `json:"items_left"`
}

type NewRolesRequest struct {
	Role_Name string `json:"role_name" validate:"required,unique"`
}

type UpdateRolesRequest struct {
	Role_Name string `json:"role_name" validate:"required,unique"`
}
type RolesService interface {
	// GetAllRoles(limit int, offset int) ([]RolesResponseAll, error)
	GetAllRoles(name []string, limit, offset int) ([]RolesResponseAll, error)
	CountRoles(name []string) (int, error)
	GetRoleById(id uint) (*RolesResponse, error)
	CreateRole(NewRolesRequest) (*RolesResponse, error)
	EditRole(id uint, newRoleName string) (*RolesResponse, error)
	DeleteRole(id uint) error
}
