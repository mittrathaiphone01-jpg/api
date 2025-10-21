package service

import (
	"fmt"
	"rrmobile/respository"
	"time"
)

type rulesService struct {
	rulesRepository respository.RulesRepository
}

func NewRulesService(rulesRepository respository.RulesRepository) RulesService {
	return &rulesService{rulesRepository: rulesRepository}
}

func (s *rulesService) GetAllRules(limit, offset int) ([]RulesResponse, error) {
	rules, err := s.rulesRepository.GetAllRules(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve rules: ", err)
	}

	var rulesResponse []RulesResponse
	for _, rule := range rules {
		rulesResponse = append(rulesResponse, RulesResponse{
			Id:               rule.Id,
			Threshold_Months: rule.Threshold_Months,
			Type_Discount:    rule.Type_Discount,
			Discount_Amount:  rule.Discount_Amount,
			CreatedAt:        rule.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        rule.UpdatedAt.Format(time.RFC3339),
		})
	}
	return rulesResponse, nil
}

func (s *rulesService) CountRules() (int64, error) {
	var count int64
	err := s.rulesRepository.CountRules(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *rulesService) CreateRule(request NewRuleRequest) (*RulesResponse, error) {
	if request.Threshold_Months <= 0 {
		err := fmt.Errorf("Threshold months must be greater than 0")
		return nil, err
	}
	if request.Discount_Amount < 0 {
		err := fmt.Errorf("Discount amount cannot be negative")
		return nil, err
	}

	rule := respository.Rules{
		Threshold_Months: request.Threshold_Months,
		Type_Discount:    request.Type_Discount,
		Discount_Amount:  request.Discount_Amount,
	}

	newRule, err := s.rulesRepository.AddRule(rule)
	if err != nil {
		return nil, fmt.Errorf("Failed to create rule: ", err)
	}

	response := RulesResponse{
		Id:               newRule.Id,
		Threshold_Months: newRule.Threshold_Months,
		Type_Discount:    newRule.Type_Discount,
		Discount_Amount:  newRule.Discount_Amount,
		CreatedAt:        newRule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        newRule.UpdatedAt.Format(time.RFC3339),
	}

	return &response, nil
}

func (s *rulesService) GetRuleByID(id uint) (*RulesResponse, error) {
	rule, err := s.rulesRepository.GetRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("Rule with ID  not found: ", id, err)
	}
	response := RulesResponse{
		Id:               rule.Id,
		Threshold_Months: rule.Threshold_Months,
		Type_Discount:    rule.Type_Discount,
		Discount_Amount:  rule.Discount_Amount,
		CreatedAt:        rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        rule.UpdatedAt.Format(time.RFC3339),
	}
	return &response, nil
}

func (s *rulesService) EditRule(id uint, request UpdateRuleRequest) (*RulesResponse, error) {
	if request.Threshold_Months <= 0 {
		return nil, fmt.Errorf("Threshold months must be greater than 0")
	}
	if request.Discount_Amount < 0 {
		return nil, fmt.Errorf("Discount amount cannot be negative")
	}

	rule := respository.Rules{
		Threshold_Months: request.Threshold_Months,
		Type_Discount:    request.Type_Discount,
		Discount_Amount:  request.Discount_Amount,
	}

	updatedRule, err := s.rulesRepository.UpdateRule(id, rule)
	if err != nil {
		return nil, fmt.Errorf("Failed to update rule with ID", err)
	}

	response := RulesResponse{
		Id:               updatedRule.Id,
		Threshold_Months: updatedRule.Threshold_Months,
		Type_Discount:    updatedRule.Type_Discount,
		Discount_Amount:  updatedRule.Discount_Amount,
		CreatedAt:        updatedRule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        updatedRule.UpdatedAt.Format(time.RFC3339),
	}

	return &response, nil
}

func (s *rulesService) DeleteRule(id uint) error {
	err := s.rulesRepository.DeleteRule(id)
	if err != nil {
		return fmt.Errorf("Failed to delete rule with ID ", id)
	}
	return nil
}
