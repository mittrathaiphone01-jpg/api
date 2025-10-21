package service

import "rrmobile/respository"

type fineCategoryService struct {
	fineCategoryRepository respository.FineCategoryRepository
}

func NewFineCategoryService(fineCategoryRepository respository.FineCategoryRepository) FineCategoryService {
	return &fineCategoryService{fineCategoryRepository: fineCategoryRepository}
}
func (s *fineCategoryService) GetFineCategories(limit, offset int) ([]FineCategoryResponse, error) {
	categories, err := s.fineCategoryRepository.GetFineCategories(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []FineCategoryResponse
	for _, category := range categories {
		responses = append(responses, FineCategoryResponse{
			Id:   category.Id,
			Name: category.Name,
		})
	}
	return responses, nil
}
func (s *fineCategoryService) GetFineCategoryById(id uint) (*FineCategoryResponse, error) {
	category, err := s.fineCategoryRepository.GetFineCategoryById(id)
	if err != nil {
		return nil, err
	}
	return &FineCategoryResponse{
		Id:   category.Id,
		Name: category.Name,
	}, nil
}
func (s *fineCategoryService) CreateFineCategory(request NewFineCategoryRequest) (*FineCategoryResponse, error) {
	category := respository.Fine_System_Category{
		Name: request.Name,
	}
	createdCategory, err := s.fineCategoryRepository.CreateFineCategory(category)
	if err != nil {
		return nil, err
	}
	return &FineCategoryResponse{
		Id:   createdCategory.Id,
		Name: createdCategory.Name,
	}, nil
}

func (s *fineCategoryService) UpdateFineCategory(id uint, request UpdateFineCategoryRequest) (*FineCategoryResponse, error) {
	category := respository.Fine_System_Category{
		Id:   id,
		Name: request.Name,
	}
	updatedCategory, err := s.fineCategoryRepository.UpdateFineCategory(category)
	if err != nil {
		return nil, err
	}
	return &FineCategoryResponse{
		Id:   updatedCategory.Id,
		Name: updatedCategory.Name,
	}, nil
}

func (s *fineCategoryService) DeleteFineCategory(id uint) error {
	return s.fineCategoryRepository.DeleteFineCategory(id)
}
func (s *fineCategoryService) CountFineCategories() (int64, error) {
	var count int64
	if err := s.fineCategoryRepository.CountFineCategories(&count); err != nil {
		return 0, err
	}
	return count, nil
}
