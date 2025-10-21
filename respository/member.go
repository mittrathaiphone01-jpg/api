package respository

type Member struct {
	Id       uint   `db:"id"`
	FullName string `db:"full_name"`
	Tel      string `db:"tel"`
	UserId   string `db:"user_id"`
}
type MemberFilter struct {
	FullName []string
	UserId   []string
}
type MemberRepository interface {
	GetMembers(filter MemberFilter, limit, offset int) ([]Member, error)
	CountMembers(filter MemberFilter) (int64, error)
	GetMemberById(id uint) (*Member, error)
	AddMember(member Member) (*Member, error)
	UpdateMember(member Member) (*Member, error)
	DeleteMember(id uint) error

	CheckUserId(userID string) (*Member, error)

	FindByTel(tel string) (*Member, error)
	IsUserIdExists(userId string) error
}
