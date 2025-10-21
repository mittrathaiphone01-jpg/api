package respository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type rulesRespositoryDB struct {
	db *gorm.DB
}

func NewRulesRepositoryDB(db *gorm.DB) RulesRepository {
	return &rulesRespositoryDB{db: db}
}
func (r *rulesRespositoryDB) GetAllRules(limit int, offset int) ([]Rules, error) {
	var rules []Rules
	query := r.db.Model(&Rules{})

	if limit <= 0 {
		limit = 10 // กำหนดค่าเริ่มต้นถ้า limit น้อยกว่าหรือเท่ากับ 0
	}
	if offset < 0 {
		offset = 0 // กำหนดค่าเริ่มต้นถ้า offset น้อยกว่า 0
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&rules).Error

	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil // คืนค่า nil ถ้าไม่มีข้อมูล
	}
	return rules, nil
}
func (r *rulesRespositoryDB) CountRules(count *int64) error {
	return r.db.Model(&Rules{}).Count(count).Error
}
func (r *rulesRespositoryDB) AddRule(rule Rules) (*Rules, error) {
	existingRule := Rules{}
	tx := r.db.Where("threshold_months = ? AND type_discount = ? AND discount_amount = ?", rule.Threshold_Months, rule.Type_Discount, rule.Discount_Amount).First(&existingRule)
	if tx.RowsAffected > 0 {
		return nil, fmt.Errorf("This rule already exists")
	}
	err := r.db.Create(&rule).Error
	if err != nil {
		return nil, err
	}

	return &rule, nil
}


func (r *rulesRespositoryDB) GetRuleByID(id uint) (*Rules, error) {
	var rule Rules
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateRule updates a rule, but prevents changes to ThresholdMonths if the new value already exists
func (r *rulesRespositoryDB) UpdateRule(id uint, newRule Rules) (*Rules, error) {
	// 1. Fetch the existing rule
	var existingRule Rules
	if err := r.db.First(&existingRule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rule not found")
		}
		return nil, err
	}

	// 2. Check if the ThresholdMonths value is being changed
	if existingRule.Threshold_Months != newRule.Threshold_Months {
		// 3. If it's being changed, check if the new value already exists in the database
		var count int64
		if err := r.db.Model(&Rules{}).Where("threshold_months = ?", newRule.Threshold_Months).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			// A record with this ThresholdMonths value already exists
			return nil, errors.New("cannot update ThresholdMonths to a value that already exists in the database")
		}
	}

	// 4. Perform the update for all fields
	if err := r.db.Model(&Rules{}).Where("id = ?", id).Updates(newRule).Error; err != nil {
		return nil, err
	}

	return &newRule, nil
}
func (r *rulesRespositoryDB) DeleteRule(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&Rules{}).Error; err != nil {
		return err
	}
	return nil
}
