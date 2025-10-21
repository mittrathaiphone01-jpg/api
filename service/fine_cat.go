package service

type FineCategoryResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type NewFineCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}
type UpdateFineCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type FineCategoryService interface {
	GetFineCategories(limit, offset int) ([]FineCategoryResponse, error)
	GetFineCategoryById(id uint) (*FineCategoryResponse, error)
	CreateFineCategory(request NewFineCategoryRequest) (*FineCategoryResponse, error)
	UpdateFineCategory(id uint, request UpdateFineCategoryRequest) (*FineCategoryResponse, error)
	DeleteFineCategory(id uint) error
	CountFineCategories() (int64, error)
}
