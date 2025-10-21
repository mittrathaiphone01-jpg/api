package respository

import "time"

type Rules struct {
	Id               uint      `db:"id"`
	Threshold_Months int       `db:"threshold_months"`
	Type_Discount    bool      `db:"type_discount"`
	Discount_Amount  float64   `db:"discount_amount"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type RulesRepository interface {
	GetAllRules(limit int, offset int) ([]Rules, error)
	CountRules(count *int64) error
	AddRule(rule Rules) (*Rules, error)
	GetRuleByID(id uint) (*Rules, error)
	UpdateRule(id uint, rule Rules) (*Rules, error)
	DeleteRule(id uint) error
}
