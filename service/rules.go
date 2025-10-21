package service

type RulesResponse struct {
	Id               uint    `json:"id"`
	Threshold_Months int     `json:"threshold_months"`
	Type_Discount    bool    `json:"type_discount"`
	Discount_Amount  float64 `json:"discount_amount"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type PaginatedRulesResponse struct {
	Rules       []RulesResponse `json:"results"`
	CurrentPage int             `json:"current_page"`
	TotalPages  int             `json:"total_pages"`
	ItemsLeft   int             `json:"items_left"`
}
type NewRuleRequest struct {
	Threshold_Months int     `json:"threshold_months" `
	Type_Discount    bool    `json:"type_discount"`
	Discount_Amount  float64 `json:"discount_amount"`
}

type UpdateRuleRequest struct {
	Threshold_Months int     `json:"threshold_months" validate:"required"`
	Type_Discount    bool    `json:"type_discount"`
	Discount_Amount  float64 `json:"discount_amount" validate:"required"`
}

type RulesService interface {
	GetAllRules(limit, offset int) ([]RulesResponse, error)
	CountRules() (int64, error)
	GetRuleByID(id uint) (*RulesResponse, error)
	CreateRule(request NewRuleRequest) (*RulesResponse, error)
	EditRule(id uint, request UpdateRuleRequest) (*RulesResponse, error)
	DeleteRule(id uint) error
}
