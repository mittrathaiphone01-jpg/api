package service

type InstallmentResponse struct {
	Id  uint `json:"id"`
	Day int  `json:"day"`
}
type NewInstallmentRequest struct {
	Day int `json:"day" validate:"required"`
}
type UpdateInstallmentRequest struct {
	Day int `json:"day" validate:"required"`
}

type InstallmentService interface {
	GetInstallments(limit, offset int) ([]InstallmentResponse, error)
	GetInstallmentById(id uint) (*InstallmentResponse, error)
	CreateInstallment(request NewInstallmentRequest) (*InstallmentResponse, error)
	UpdateInstallment(id uint, request UpdateInstallmentRequest) (*InstallmentResponse, error)
	DeleteInstallment(id uint) error
}
