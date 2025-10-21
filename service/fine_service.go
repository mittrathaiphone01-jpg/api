package service

import (
	"fmt"
	"rrmobile/respository"
)

type fineService struct {
	fineRepository respository.FineRepository
}

func NewFineService(fineRepository respository.FineRepository) FineService {
	return &fineService{fineRepository: fineRepository}
}

func (s *fineService) GetFines(limit, offset int) ([]FineResponse, error) {
	fines, err := s.fineRepository.GetFines(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []FineResponse
	for _, fine := range fines {
		responses = append(responses, FineResponse{
			Id:             fine.Id,
			FineAmount:     fine.FineAmount,
			FineCategoryId: fine.Fine_System_CategoryId,
		})
	}
	return responses, nil
}
func (s *fineService) GetFineById(id uint) (*FineResponse, error) {
	fine, err := s.fineRepository.GetFineById(id)
	if err != nil {
		return nil, err
	}
	return &FineResponse{
		Id:             fine.Id,
		FineAmount:     fine.FineAmount,
		FineCategoryId: fine.Fine_System_CategoryId,
	}, nil
}

func (s *fineService) CreateFine(request NewFineRequest) (*FineResponse, error) {
	fine := respository.Fine_System{
		FineAmount:             request.FineAmount,
		Fine_System_CategoryId: request.FineCategoryId,
	}
	if fine.FineAmount <= 0 {
		return nil, fmt.Errorf("จำนวนค่าปรับต้องมากกว่า 0")
	}
	createdFine, err := s.fineRepository.CreateFine(fine)
	if err != nil {
		return nil, err
	}
	return &FineResponse{
		Id:             createdFine.Id,
		FineAmount:     createdFine.FineAmount,
		FineCategoryId: createdFine.Fine_System_CategoryId,
	}, nil
}
func (s *fineService) UpdateFine(id uint, request UpdateFineRequest) (*FineResponse, error) {
	fine := respository.Fine_System{
		Id:                     id,
		FineAmount:             request.FineAmount,
		Fine_System_CategoryId: request.FineCategoryId,
	}
	updatedFine, err := s.fineRepository.UpdateFine(fine)
	if err != nil {
		return nil, err
	}
	return &FineResponse{
		Id:             updatedFine.Id,
		FineAmount:     updatedFine.FineAmount,
		FineCategoryId: updatedFine.Fine_System_CategoryId,
	}, nil
}

func (s *fineService) DeleteFine(id uint) error {
	return s.fineRepository.DeleteFine(id)
}
func (s *fineService) CountFines() (int64, error) {
	var count int64
	if err := s.fineRepository.CountFines(&count); err != nil {
		return 0, err
	}
	return count, nil
}
