package respository

import (
	"rrmobile/model"
	"time"
)

type Bill_Header struct {
	Id        uint      `db:"id"`
	Invoice   string    `db:"invoice"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
	MemberId  uint      `db:"member_id"`
	User_Id   int       `db:"user_id"`
	ProductId uint      `db:"product_id"`
	//เหลือทำต่อหลังจาก Product
	Extra_Percent      int     `db:"extra_percent"`
	Down_Percent       int     `db:"down_percent"`
	Installments_Month int     `db:"installments_month"`
	Net_installment    float64 `db:"net_installment"`

	Total_Price      float64 `db:"total_price"`
	Paid_Amount      int     `db:"paid_amount"`
	Remaining_Amount float64 `db:"remaining_amount"`

	Total_Installments     int `db:"total_installments"`
	Paid_Installments      int `db:"paid_installments"`
	Remaining_Installments int `db:"remaining_installments"`

	Late_Day   int     `db:"late_day"`
	Fee_Amount float64 `db:"fee_amount"`

	Status int    `db:"status"`
	Note   string `db:"note"`

	Credit_Balance float64        `db:"credit_balance"`
	BillDetails    []Bill_Details `db:"bill_details"`
	Member         Member         `db:"members"`
	Product        Product        `db:"products"`
	User           User           `db:"users"`
}

type Bill_Details struct {
	Id                uint    `db:"id"`
	Bill_HeaderId     uint    `db:"bill_header_id"`
	Installment_Price float64 `db:"installment_price"`
	Paid_Amount       float64 `db:"paid_amount"`

	Payment_Date time.Time `db:"payment_date"`
	UpdatedAt    time.Time `db:"updated_at"`

	Fee_Amount float64 `db:"fee_amount"`
	Status     int     `db:"status"`

	Credit_Balance float64 `db:"credit_balance"`
	Payment_No     string  `db:"payment_no"`
}
type Bill_Details1 struct {
	Id                uint    `db:"id"`
	Bill_HeaderId     uint    `db:"bill_header_id"`
	Installment_Price float64 `db:"installment_price"`
	Paid_Amount       float64 `db:"paid_amount"`

	Payment_Date time.Time `db:"payment_date"`
	UpdatedAt    time.Time `db:"updated_at"`

	Fee_Amount float64 `db:"fee_amount"`
	Status     int     `db:"status"`

	Credit_Balance float64     `db:"credit_balance"`
	BillHeader     Bill_Header `db:"bill_header"`
	Payment_No     string      `db:"payment_no"`
}

type BillFilter struct {
	Invs         []string
	Names        []string
	CategoryID   *uint
	IsActive     *bool
	DateFrom     *time.Time
	DateTo       *time.Time
	SortPrice    string // "asc" หรือ "desc"
	NameOrPhones []string
}

type BestSellingProduct struct {
	ProductId  uint
	TotalSales int64
}

type BillSummary struct {
	PaidTotal   float64
	UnpaidTotal float64
	PaidCount   int64
	UnpaidCount int64
}
type BillRepository interface {
	CreateBill(bill *Bill_Header) (*Bill_Header, error)
	GetBillById(id uint) (*Bill_Header, error)
	CreateBillDetails(details []Bill_Details) error
	GetUnpaidInstallments(billID uint) ([]Bill_Details, error)
	GetPaidInstallments(billID uint) ([]Bill_Details, error)
	UpdateBillDetail(installments []Bill_Details) error
	GetBillDetailById(id uint) (*Bill_Details, error)
	UpdateBill(bill *Bill_Header) error
	UpdateSingleBillDetail(detail *Bill_Details) error
	UpdateBillFee(bill *Bill_Header) error // UpdateBillHeader(bill Bill_Header) error
	GetAllBillDetails(billID uint) ([]Bill_Details, error)
	GetAllUnpaidBills() ([]model.Bill_Header, error)
	GetAllInstallments(billID uint) ([]Bill_Details, error)
	GetBillWithUnpaidInstallments(billID uint) (*Bill_Header, error)
	GetUnpaidBillsBatch(limit, offset int) ([]model.Bill_Header, error)
	GetUnpaidInstallmentsByBillIDs(billIDs []uint) ([]model.Bill_Details, error)
	UpdateBillDetailsBatch(details []model.Bill_Details) error
	UpdateBillsBatch(bills []*model.Bill_Header) error
	UpdateBillDetailBatch(installments []model.Bill_Details) error
	UpdateBillBatch(bills []*model.Bill_Header) error
	GetLastInvByYear(yearSuffix string) (string, error)
	GetAllBill(filter BillFilter, limit, offset int, bestProductIds []uint, sortOrder int) ([]Bill_Header, error)
	CountBills(filter BillFilter) (int64, error)
	GetBestSellingProducts(limit int) ([]BestSellingProduct, error)
	GetLastHpcByYear(yearSuffix string) (string, error)
	CreateInstallmentBill(bill *model.Bill_Header_Installment) (*model.Bill_Header_Installment, error)
	CreateInstallmentBillDetails(details []model.Bill_Details_Installment) error
	GetUnpaidBillInstallments(billID uint) ([]model.Bill_Details_Installment, error)
	UpdateInstallmentBillDetail(installments []model.Bill_Details_Installment) error
	UpdateBillInstallment(bill *model.Bill_Header_Installment) error
	GetInstallmentBillById(id uint) (*model.Bill_Header_Installment, error)
	GetInstallmentDetailsByBillID(billID uint) ([]model.Bill_Details_Installment, error)

	GetPaidInstallmentsBill(billID uint) ([]model.Bill_Details_Installment, error)
	GetAllInstallmentUnpaidBills() ([]model.Bill_Header_Installment, error)

	UpdateInstallmentBill(bill *model.Bill_Header_Installment) error
	UpdateBillFeeInstallment(bill *model.Bill_Header_Installment) error
	UpdateBillInstallmentDetail(installments []model.Bill_Details_Installment) error
	GetInstallmentBillDetailById(id uint) (*model.Bill_Details_Installment, error)
	UpdateInstallmentSingleBillDetail(detail *model.Bill_Details_Installment) error

	GetInstallmentAllBill(
		filter BillFilter,
		limit, offset int,
		bestProductIds []uint,
		sortOrder int,
	) ([]model.Bill_Header_Installment, error)
	CountInstallmentBills(filter BillFilter) (int64, error)
	GetBestSellingInstallmentsProducts(limit int) ([]BestSellingProduct, error)

	SumPaidAndUnpaidCounts(filter BillFilter) (paidCount int64, unpaidCount int64, err error)
	GetBillSummary(filter BillFilter) (summary BillSummary, err error)
	SumInstallmentPaidAndUnpaidCounts(filter BillFilter) (paidCount int64, unpaidCount int64, err error)
	GetInstallmentBillSummary(filter BillFilter) (summary BillSummary, err error)

	UpdateBill_Installments(bill *model.Bill_Header_Installment) error

	GetUnPayAllBill(filter BillFilter, limit, offset int, sortOrder int) ([]Bill_Header, error)
	CountUnpayBills(filter BillFilter) (int64, error)
	GetInstallmentAllBillUnpay(
		filter BillFilter,
		limit, offset int,
		sortOrder int,
	) ([]model.Bill_Header_Installment, error)
	CountInstallmentBillsUnpay(filter BillFilter) (int64, error)
	UpdateBillStatus(bill *Bill_Header) error
	UpdateInstallmentBillStatus(bill *model.Bill_Header_Installment) error

	GetUnpaidInstallmentsByDate() ([]Bill_Details1, error)
	GetUnpaidInstallBillmentsByDate() ([]model.Bill_Details_Installment, error)

	GetUnpaidBill(userId string) ([]Bill_Details1, error)

	GetUnpaidInstallmentBill(userId string) ([]model.Bill_Details_Installment, error)

	GetUnpaidInstallments1(billID uint, detailID uint) ([]Bill_Details, error)
	GetPaidInstallments1(billID uint) ([]Bill_Details, error)
	GetInstallmentCounts(billID uint) (int64, int64, error)
	GetUnpaidBillInstallments1(billID uint, detailID uint) ([]model.Bill_Details_Installment, error)
	GetPaidInstallmentsBill1(billID uint) ([]model.Bill_Details_Installment, error)

	GetpaidBill(billID uint, detailID uint) ([]Bill_Details1, error)
	GetpaidInstallmentBill(billID uint, detailID uint) ([]model.Bill_Details_Installment, error)

	GetUnpaidInstallments2(billID uint) ([]Bill_Details, error)
	GetUnpaidBillInstallments2(billID uint) ([]model.Bill_Details_Installment, error)

	SumPaidAmountByStatus1(filter BillFilter) (float64, error)
	SumPaidAmountByInstallmentStatus1(filter BillFilter) (float64, error)
	GetAllUnpaid10DayBills() ([]model.Bill_Header_Installment, error)
	UpdateInstallmentBillDetail1(installment *model.Bill_Details_Installment) error

	SumPaidAmountByStatus2(filter BillFilter) (float64, error)
	SumFeeByStatus2(filter BillFilter) (float64, error)

	SumPaidAmountByInstallmentStatus2(filter BillFilter) (float64, error)
	SumFeeInstallmentByStatus2(filter BillFilter) (float64, error)
	GetBillHeaderById(id uint) (*Bill_Header, error)

	CreateInstallmentDetail1(detail *model.Bill_Details_Installment) error
	UpdateInstallmentDetail(detail *model.Bill_Details_Installment) error

	HasInterestInPeriod(billID uint, lastRenew, nextDue time.Time) (bool, error)
	CloseOldInterestInstallments(billID uint) error


	GetUnpaidBillInstallments3(billID uint) ([]model.Bill_Details_Installment, error)
}
