package respository

type Roles struct {
	Id        uint   `db:"id"`
	Role_Name string `db:"role_name"`
}

type RoleRepository interface {
	GetAllRoles(name []string, limit int, offset int) ([]Roles, error)
	CountRoles(name []string, count *int64) error
	AddRole(role Roles) (*Roles, error)
	GetRoleByID(id uint) (*Roles, error)
	UpdateRole(id uint, role Roles) (*Roles, error)
	DeleteRole(id uint) error
}
