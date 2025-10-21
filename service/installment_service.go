package service

import (
	"fmt"
	"rrmobile/respository"
)

type installmentService struct {
	installmentRepository respository.InstallmentRepository
}

func NewInstallmentService(installmentRepository respository.InstallmentRepository) InstallmentService {
	return &installmentService{installmentRepository: installmentRepository}
}

func (s *installmentService) GetInstallments(limit, offset int) ([]InstallmentResponse, error) {
	installments, err := s.installmentRepository.GetInstallments(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []InstallmentResponse
	for _, installment := range installments {
		responses = append(responses, InstallmentResponse{
			Id:  installment.Id,
			Day: installment.Day,
		})
	}
	return responses, nil
}
func (s *installmentService) GetInstallmentById(id uint) (*InstallmentResponse, error) {
	installment, err := s.installmentRepository.GetInstallmentById(id)
	if err != nil {
		return nil, err
	}
	return &InstallmentResponse{
		Id:  installment.Id,
		Day: installment.Day,
	}, nil
}
func (s *installmentService) CreateInstallment(request NewInstallmentRequest) (*InstallmentResponse, error) {
	installment := respository.Installment{
		Day: request.Day,
	}
	if installment.Day < 0 {
		return nil, fmt.Errorf("จำนวนต้องมากกว่า 0")
	}
	createdInstallment, err := s.installmentRepository.CreateInstallment(installment)
	if err != nil {
		return nil, err
	}
	return &InstallmentResponse{
		Id:  createdInstallment.Id,
		Day: createdInstallment.Day,
	}, nil
}
func (s *installmentService) UpdateInstallment(id uint, request UpdateInstallmentRequest) (*InstallmentResponse, error) {
	installment := respository.Installment{
		Id:  id,
		Day: request.Day,
	}
	updatedInstallment, err := s.installmentRepository.UpdateInstallment(installment)
	if err != nil {
		return nil, err
	}

	return &InstallmentResponse{
		Id:  updatedInstallment.Id,
		Day: updatedInstallment.Day,
	}, nil
}
func (s *installmentService) DeleteInstallment(id uint) error {
	return s.installmentRepository.DeleteInstallment(id)
}
