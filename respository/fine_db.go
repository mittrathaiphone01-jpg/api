package respository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type fineRepositoryDB struct {
	db *gorm.DB
}

func NewFineRepositoryDB(db *gorm.DB) FineRepository {
	return &fineRepositoryDB{db: db}
}

func (r *fineRepositoryDB) GetFineById(id uint) (*Fine_System, error) {
	var fine Fine_System
	if err := r.db.First(&fine, id).Error; err != nil {
		return nil, err
	}
	return &fine, nil
}

func (r *fineRepositoryDB) CreateFine(fine Fine_System) (*Fine_System, error) {
	// ตรวจสอบซ้ำใน DB ก่อน
	existingFine := Fine_System{}
	tx := r.db.Where("fine_amount = ? OR fine_system_category_id = ?",
		fine.FineAmount, fine.Fine_System_CategoryId).First(&existingFine)
	if tx.RowsAffected > 0 {
		return nil, fmt.Errorf("fine amount %v in category %v already exists",
			fine.FineAmount, fine.Fine_System_CategoryId)
	}

	// สร้าง record ใหม่
	if err := r.db.Create(&fine).Error; err != nil {
		// ถ้าเป็น duplicate จาก DB constraint ก็จับเป็น error ได้
		return nil, fmt.Errorf("failed to create fine: %w", err)
	}

	return &fine, nil
}

func (r *fineRepositoryDB) UpdateFine(fine Fine_System) (*Fine_System, error) {
	var existingFine Fine_System
	if err := r.db.First(&existingFine, fine.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("installment not found")
		}
		return nil, err
	}

	// 2. Check if Day value is being changed
	if existingFine.FineAmount != fine.FineAmount {
		// 3. If it's being changed, check if the new value already exists in the database
		var count int64
		if err := r.db.Model(&Fine_System{}).Where("fine_amount = ?", fine.FineAmount).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, errors.New("cannot update Day to a value that already exists in the database")
		}
	}

	// 4. Perform the update for all fields
	if err := r.db.Model(&Fine_System{}).Where("id = ?", fine.Id).Updates(fine).Error; err != nil {
		return nil, err
	}

	return &fine, nil
	
}
func (r *fineRepositoryDB) DeleteFine(id uint) error {
	return r.db.Delete(&Fine_System{}, id).Error
}
func (r *fineRepositoryDB) GetFines(limit int, offset int) ([]Fine_System, error) {
	var fines []Fine_System
	if err := r.db.Limit(limit).Offset(offset).Find(&fines).Error; err != nil {
		return nil, err
	}
	return fines, nil
}
func (r *fineRepositoryDB) CountFines(count *int64) error {
	return r.db.Model(&Fine_System{}).Count(count).Error
}
