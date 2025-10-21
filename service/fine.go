package service

type FineResponse struct {
	Id             uint    `json:"id"`
	FineAmount     float64 `json:"fine_amount"`
	FineCategoryId uint    `json:"fine_system_category_id"`
}
type NewFineRequest struct {
	FineAmount     float64 `json:"fine_amount" validate:"required"`
	FineCategoryId uint    `json:"fine_system_category_id" validate:"required"`
}
type UpdateFineRequest struct {
	FineAmount     float64 `json:"fine_amount" validate:"required"`
	FineCategoryId uint    `json:"fine_system_category_id" validate:"required"`
}
type FineService interface {
	GetFines(limit, offset int) ([]FineResponse, error)
	GetFineById(id uint) (*FineResponse, error)
	CreateFine(request NewFineRequest) (*FineResponse, error)
	UpdateFine(id uint, request UpdateFineRequest) (*FineResponse, error)
	DeleteFine(id uint) error
	CountFines() (int64, error)
}
