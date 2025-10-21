package respository

type Installment struct {
	Id  uint `db:"id"`
	Day int  `db:"day"`
}

type InstallmentRepository interface {
	GetInstallments(limit int, offset int) ([]Installment, error)
	GetInstallmentById(id uint) (*Installment, error)
	CreateInstallment(installment Installment) (*Installment, error)
	UpdateInstallment(installment Installment) (*Installment, error)
	DeleteInstallment(id uint) error
}
