package respository

type Fine_System struct {
	Id                     uint    `db:"id"`
	FineAmount             float64 `db:"fine"`
	Fine_System_CategoryId uint    `db:"fine_system_category_id"`
}

type FineRepository interface {
	GetFineById(id uint) (*Fine_System, error)
	CreateFine(fine Fine_System) (*Fine_System, error)
	UpdateFine(fine Fine_System) (*Fine_System, error)
	DeleteFine(id uint) error
	GetFines(limit int, offset int) ([]Fine_System, error)
	CountFines(count *int64) error
}
