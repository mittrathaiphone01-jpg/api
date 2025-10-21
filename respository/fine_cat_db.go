package respository

import (
	"errors"

	"gorm.io/gorm"
)

type FineCategoryRepositoryDB struct {
	db *gorm.DB
}

func NewFineCategoryRepositoryDB(db *gorm.DB) FineCategoryRepository {
	return &FineCategoryRepositoryDB{db: db}
}
func (r *FineCategoryRepositoryDB) GetFineCategoryById(id uint) (*Fine_System_Category, error) {
	var category Fine_System_Category
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
func (r *FineCategoryRepositoryDB) CreateFineCategory(category Fine_System_Category) (*Fine_System_Category, error) {
	existingCategory := Fine_System_Category{}
	tx := r.db.Where("name = ?", category.Name).First(&existingCategory)
	if tx.RowsAffected > 0 {
		return nil, gorm.ErrDuplicatedKey
	}
	if err := r.db.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *FineCategoryRepositoryDB) UpdateFineCategory(category Fine_System_Category) (*Fine_System_Category, error) {
	var existingCategory Fine_System_Category
	if err := r.db.First(&existingCategory, category.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("fine category not found")
		}
		return nil, err
	}

	if existingCategory.Name != category.Name {
		var count int64
		if err := r.db.Model(&Fine_System_Category{}).Where("name = ?", category.Name).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, errors.New("cannot update name to a value that already exists in the database")
		}
	}

	if err := r.db.Save(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *FineCategoryRepositoryDB) DeleteFineCategory(id uint) error {
	return r.db.Delete(&Fine_System_Category{}, id).Error
}

func (r *FineCategoryRepositoryDB) GetFineCategories(limit int, offset int) ([]Fine_System_Category, error) {
	var categories []Fine_System_Category
	if err := r.db.Limit(limit).Offset(offset).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
func (r *FineCategoryRepositoryDB) CountFineCategories(count *int64) error {
	return r.db.Model(&Fine_System_Category{}).Count(count).Error
}
