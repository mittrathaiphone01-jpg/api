package service

type MemberResponse struct {
	Id       uint   `json:"id"`
	FullName string `json:"full_name"`
	Tel      string `json:"tel"`
	// UserId   string `json:"user_id"`
}

type PaginatedMemberResponse struct {
	Member      []MemberResponse `json:"data"`
	CurrentPage int              `json:"current_page"`
	TotalPages  int              `json:"total_pages"`
	ItemsLeft   int              `json:"items_left"`
}
type PaginationResponseMember struct {
	Total       int64            `json:"total"`
	TotalPages  int              `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	HasNext     bool             `json:"has_next"`
	HasPrev     bool             `json:"has_prev"`
	Limit       int              `json:"limit"`
	Member      []MemberResponse `json:"data"`
}
type NewMemberRequest struct {
	FullName string `json:"full_name" validate:"required,unique"`
	Tel      string `json:"tel"`
	UserId   string `json:"user_id"`
}
type UpdateMemberRequest struct {
	FullName string  `json:"full_name" validate:"required,unique"`
	Tel      *string `json:"tel"`
	UserId   *string `json:"user_id"`
}
type UpdateMemberRequest1 struct {
	Tel    *string `json:"tel"`     // ใช้ pointer เพื่อรู้ว่า ส่งมาหรือไม่
	UserId *string `json:"user_id"` // ใช้ pointer เช่นกัน
}

type MemberService interface {
	GetAllMembers(
		fullname, user_id []string,
		page, limit int,
	) (*PaginationResponseMember, error)
	GetMemberById(id uint) (*MemberResponse, error)
	CreateMember(req NewMemberRequest) (*MemberResponse, error)
	EditMember(id uint, req UpdateMemberRequest) (*MemberResponse, error)
	DeleteMember(id uint) error

	GetMemberByUserId(userID string) (*MemberResponse, error)
	UpdateMemberByTel(req UpdateMemberRequest1) (*MemberResponse, error)
}

