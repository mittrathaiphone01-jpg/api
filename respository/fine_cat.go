package respository

type Fine_System_Category struct {
	Id   uint   `db:"id"`
	Name string `db:"name"`
}

type FineCategoryRepository interface {
	GetFineCategoryById(id uint) (*Fine_System_Category, error)
	CreateFineCategory(category Fine_System_Category) (*Fine_System_Category, error)
	UpdateFineCategory(category Fine_System_Category) (*Fine_System_Category, error)
	DeleteFineCategory(id uint) error
	GetFineCategories(limit int, offset int) ([]Fine_System_Category, error)
	CountFineCategories(count *int64) error
}
