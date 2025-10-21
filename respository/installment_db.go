package respository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type installmentRepositoryDB struct {
	db *gorm.DB
}

func NewInstallmentRepositoryDB(db *gorm.DB) InstallmentRepository {
	return &installmentRepositoryDB{db: db}
}
func (r *installmentRepositoryDB) GetInstallments(limit int, offset int) ([]Installment, error) {
	var installments []Installment
	if err := r.db.Limit(limit).Offset(offset).Find(&installments).Error; err != nil {
		return nil, err
	}
	return installments, nil
}
func (r *installmentRepositoryDB) GetInstallmentById(id uint) (*Installment, error) {
	var installment Installment
	if err := r.db.Find(&installment, id).Error; err != nil {
		return nil, err
	}
	return &installment, nil
}
func (r *installmentRepositoryDB) CreateInstallment(installment Installment) (*Installment, error) {
	existingInstallment := Installment{}
	tx := r.db.Where("day = ?", installment.Day).First(&existingInstallment)
	if tx.RowsAffected > 0 {
		return nil, fmt.Errorf("Installment with this day already exists")
	}
	if err := r.db.Create(&installment).Error; err != nil {
		return nil, err
	}
	return &installment, nil

}
func (r *installmentRepositoryDB) UpdateInstallment(installment Installment) (*Installment, error) {
	// 1. Fetch the existing installment
	var existingInstallment Installment
	if err := r.db.First(&existingInstallment, installment.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("installment not found")
		}
		return nil, err
	}

	// 2. Check if Day value is being changed
	if existingInstallment.Day != installment.Day {
		// 3. If it's being changed, check if the new value already exists in the database
		var count int64
		if err := r.db.Model(&Installment{}).Where("day = ?", installment.Day).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, errors.New("cannot update Day to a value that already exists in the database")
		}
	}

	// 4. Perform the update for all fields
	if err := r.db.Model(&Installment{}).Where("id = ?", installment.Id).Updates(installment).Error; err != nil {
		return nil, err
	}

	return &installment, nil
}

// 	// 1. Fetch the existing rule

func (r *installmentRepositoryDB) DeleteInstallment(id uint) error {
	if err := r.db.Delete(&Installment{}, id).Error; err != nil {
		return err
	}
	return nil
}
