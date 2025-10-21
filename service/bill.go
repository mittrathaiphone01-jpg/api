package service

import (
	"rrmobile/model"
	"time"
)

type Bill_HeaderResponse struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status   int     `json:"status"`
	Note     string  `json:"note"`
	TermType int     `json:"term_type"`
	SumPaid  float64 `json:"sum_paid_amount"` // ✅ เพิ่มฟิลด์ใหม่

	Credit_Balance float64                `json:"credit_balance"`
	BillDetails    []Bill_DetailsResponse `json:"bill_details"`
}

type Bill_DetailsResponse struct {
	Id                uint    `json:"id"`
	Bill_HeaderId     uint    `json:"bill_header_id"`
	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`

	Fee_Amount float64 `json:"fee_amount"`
	Status     int     `json:"status"`

	Credit_Balance float64 `json:"credit_balance"`
	Payment_No     string  `json:"payment_no"`
}

type Bill_HeaderResponse1 struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status   int    `json:"status"`
	Note     string `json:"note"`
	TermType int    `json:"term_type"`

	Credit_Balance float64                 `json:"credit_balance"`
	BillDetails    []Bill_DetailsResponse1 `json:"data"` // ✅ slice
}

type Bill_DetailsResponse1 struct {
	Id                uint    `json:"id"`
	Bill_HeaderId     uint    `json:"bill_header_id"`
	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`

	Fee_Amount     float64              `json:"fee_amount"`
	Status         int                  `json:"status"`
	Credit_Balance float64              `json:"credit_balance"`
	Bill_Header    *Bill_HeaderResponse `json:"bill_header"` // ✅ เป็น pointer ไม่ใช่ slice
	Payment_No     string               `json:"payment_no"`
}
type Bill_HeaderResponse2 struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`
	MemberUser_id  string    `json:"member_user_id"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status   int    `json:"status"`
	Note     string `json:"note"`
	TermType int    `json:"term_type"`

	Credit_Balance float64                 `json:"credit_balance"`
	BillDetails    []Bill_DetailsResponse2 `json:"data"` // ✅ slice

}

type Bill_DetailsResponse2 struct {
	Id                uint    `json:"id"`
	Bill_HeaderId     uint    `json:"bill_header_id"`
	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`

	Fee_Amount     float64                `json:"fee_amount"`
	Status         int                    `json:"status"`
	Credit_Balance float64                `json:"credit_balance"`
	Bill_Header    []Bill_HeaderResponse2 `json:"data"` // ✅ เป็น pointer ไม่ใช่ slice
	Payment_No     string                 `json:"payment_no"`
}
type NewBillHeader struct {
	Invoice            string  `json:"invoice"`
	MemberId           uint    `json:"member_id"`
	User_Id            int     `json:"user_id"`
	ProductId          uint    `json:"product_id"`
	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`
	Total_Price        float64 `json:"total_price"`
	Paid_Amount        int     `json:"paid_amount"`
	Remaining_Amount   float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status int `json:"status"`
}

type UpdateAddExtraRequest struct {
	BillID        uint    `json:"bill_id"`
	InstallmentID uint    `json:"installment_id"`
	Paid_Amount   float64 `json:"paid_amount"`
}
type PaginationResponseBill struct {
	Total       int64                 `json:"total"`
	TotalPages  int                   `json:"total_pages"`
	CurrentPage int                   `json:"current_page"`
	HasNext     bool                  `json:"has_next"`
	HasPrev     bool                  `json:"has_prev"`
	Limit       int                   `json:"limit"`
	Header      BillHeaderSummary     `json:"header"` // ✅ เพิ่มตรงนี้
	Bills       []Bill_HeaderResponse `json:"data"`
	SumPaid     float64               `json:"sum_paid_amount"` // ✅ เพิ่มฟิลด์ใหม่
	FeeAmount   float64               `json:"sum_fee_amount"`
	SumUnpaid   float64               `json:"sum_unpaid_amount"`
}
type PaginationResponseBillInstallment struct {
	Total       int64                             `json:"total"`
	TotalPages  int                               `json:"total_pages"`
	CurrentPage int                               `json:"current_page"`
	HasNext     bool                              `json:"has_next"`
	HasPrev     bool                              `json:"has_prev"`
	Limit       int                               `json:"limit"`
	Header      BillHeaderInstallmentSummary      `json:"header"` // ✅ เพิ่มตรงนี้
	Bills       []Bill_HeaderResponse_Installment `json:"data"`
	SumPaid     float64                           `json:"sum_paid_amount"` // ✅ เพิ่มฟิลด์ใหม่
	FeeAmount   float64                           `json:"sum_fee_amount"`
	SumUnpaid   float64                           `json:"sum_unpaid_amount"`
}

type NewInstallmentBillHeader struct {
	Invoice   string `json:"invoice"`
	MemberId  uint   `json:"member_id"`
	User_Id   int    `json:"user_id"`
	ProductId uint   `json:"product_id"`
	// Installment_Day int    `json:"installments_day"`
	InstallmentId    uint    `json:"installment_id"` // 👈 เพิ่มอันนี้
	Installment_Day  int     `json:"installment_day"`
	Extra_Percent    int     `json:"extra_percent"`
	Net_installment  float64 `json:"net_installment"`
	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount int     `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status int `json:"status"`

	TermValue             int     `json:"term_value"`
	Loan_Amount           float64 `json:"loan_amount"`
	Interest_Amount       float64 `json:"interest_amount"`
	Total_Interest_Amount float64 `json:"total_interest_amount"`

	// Installmen	TermType        int
	TermType int `json:"term_type"`
}

type Bill_HeaderResponse_Installment struct {
	Id             uint      `json:"id"`
	Invoice        string    `json:"invoice"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status                int     `json:"status"`
	Note                  string  `json:"note"`
	TermType              int     `json:"term_type"`
	Interest_Amount       float64 `json:"interest_amount"`
	Total_Interest_Amount float64 `json:"total_interest_amount"`
	Loan_Amount           float64 `json:"loan_amount"`

	Credit_Balance float64                `json:"credit_balance"`
	BillDetails    []Bill_DetailsResponse `json:"bill_details"`
}
type UpdateAddExtraRequest_Installment struct {
	BillID        uint    `json:"bill_id"`
	InstallmentID uint    `json:"installment_id"`
	Paid_Amount   float64 `json:"paid_amount"`
}
type BillHeaderSummary struct {
	PaidBillCount   int64 `json:"paid_bill_count"`
	UnpaidBillCount int64 `json:"unpaid_bill_count"`
}
type BillHeaderInstallmentSummary struct {
	PaidBillCount   int64 `json:"paid_bill_count"`
	UnpaidBillCount int64 `json:"unpaid_bill_count"`
}
type Update_Installment struct {
	Status int    `json:"status"`
	Note   string `json:"note"`
}
type InstallmentPayResult struct {
	InstallmentNo int     `json:"installment_no"`
	Case          string  `json:"case"`
	Message       string  `json:"message"`
	CreditLeft    float64 `json:"credit_left"`
	PaidAmount    float64 `json:"paid_amount"`
}
type Bill_Details_Installment struct {
	Id uint `json:"id"`

	Bill_Header_InstallmentId uint `json:"bill_header_installment_id"`

	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`

	Fee_Amount float64 `json:"fee_amount"`
	Status     int     `json:"status"`

	Credit_Balance float64 `json:"credit_balance"`
	Payment_No     string  `json:"payment_no"`
}

type Bill_HeaderResponse_Installment1 struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status                int                         `json:"status"`
	Note                  string                      `json:"note"`
	TermType              int                         `json:"term_type"`
	Interest_Amount       float64                     `json:"interest_amount"`
	Total_Interest_Amount float64                     `json:"total_interest_amount"`
	Loan_Amount           float64                     `json:"loan_amount"`
	Credit_Balance        float64                     `json:"credit_balance"`
	BillDetails           []Bill_Details_Installment1 `json:"bill_details"`
}
type Bill_Details_Installment1 struct {
	Id uint `json:"id"`

	Bill_Header_InstallmentId uint `json:"bill_header_installment_id"`

	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`

	Fee_Amount float64 `json:"fee_amount"`
	Status     int     `json:"status"`

	Credit_Balance float64                          `json:"credit_balance"`
	Bill_Header    *Bill_HeaderResponse_Installment `json:"bill_header_installments"` // ✅ เป็น pointer ไม่ใช่ slice
	Payment_No     string                           `json:"payment_no"`
}

type Close_Installment struct {
	BillID        uint    `json:"bill_id"`
	InstallmentID uint    `json:"installment_id"`
	Paid_Amount   float64 `json:"paid_amount"`
}
type Close_Bill struct {
	BillID        uint    `json:"bill_id"`
	InstallmentID uint    `json:"installment_id"`
	Paid_Amount   float64 `json:"paid_amount"`
}

type BillResponseWrapperMap struct {
	Pagination Pagination             `json:"pagination"`
	Results    map[string]interface{} `json:"results"`
}
type Pagination struct {
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
	Limit       int   `json:"limit"`
}
type BillResponseWrapperMapInstall struct {
	Pagination PaginationInstall      `json:"pagination"`
	Results    map[string]interface{} `json:"results"`
}
type PaginationInstall struct {
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
	Limit       int   `json:"limit"`
}

type Bill_HeaderResponse_Installment2 struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`
	MemberUser_id  string    `json:"member_user_id"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status                int     `json:"status"`
	Note                  string  `json:"note"`
	TermType              int     `json:"term_type"`
	Interest_Amount       float64 `json:"interest_amount"`
	Total_Interest_Amount float64 `json:"total_interest_amount"`
	Loan_Amount           float64 `json:"loan_amount"`

	Credit_Balance float64                     `json:"credit_balance"`
	BillDetails    []Bill_Details_Installment1 `json:"bill_details"`
}
type Bill_Details_Installment2 struct {
	Id uint `json:"id"`

	Bill_Header_InstallmentId uint `json:"bill_header_installment_id"`

	Installment_Price float64 `json:"installment_price"`
	Paid_Amount       float64 `json:"paid_amount"`

	Payment_Date time.Time `json:"payment_date"`
	UpdatedAt    time.Time `json:"updated_at"`

	Fee_Amount float64 `json:"fee_amount"`
	Status     int     `json:"status"`

	Credit_Balance float64                            `json:"credit_balance"`
	Bill_Header    []Bill_HeaderResponse_Installment2 `json:"data"` // ✅ เป็น pointer ไม่ใช่ slice
	Payment_No     string                             `json:"payment_no"`
}
type Bill_HeaderResponse3 struct {
	Id        uint      `json:"id"`
	Invoice   string    `json:"invoice"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	MemberId       uint   `json:"member_id"`
	MemberFullName string `json:"member_full_name"`
	MemberUser_id  string `json:"member_user_id"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price            float64 `json:"total_price"`
	Paid_Amount            int     `json:"paid_amount"`
	Remaining_Amount       float64 `json:"remaining_amount"`
	Total_Installments     int     `json:"total_installments"`
	Paid_Installments      int     `json:"paid_installments"`
	Remaining_Installments int     `json:"remaining_installments"`

	Late_Day       int     `json:"late_day"`
	Fee_Amount     float64 `json:"fee_amount"`
	Status         int     `json:"status"`
	Credit_Balance float64 `json:"credit_balance"`
	Note           string  `json:"note"`
	TermType       int     `json:"term_type"`

	Data []Bill_DetailsResponse2 `json:"data"` // 👈 ต้องมี

}

type Bill_HeaderResponse_Installment3 struct {
	Id      uint   `json:"id"`
	Invoice string `json:"invoice"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
	MemberId       uint      `json:"member_id"`
	MemberFullName string    `json:"member_full_name"`
	MemberUser_id  string    `json:"member_user_id"`

	User_Id      int    `json:"user_id"`
	UserFullName string `json:"user_full_name"`
	UserUsername string `json:"user_username"`

	ProductId       uint    `json:"product_id"`
	ProductSku      string  `json:"product_sku"`
	ProductName     string  `json:"product_name"`
	ProductPrice    float64 `json:"product_price"`
	ProductCategory string  `json:"product_category"`

	Extra_Percent      int     `json:"extra_percent"`
	Down_Percent       int     `json:"down_percent"`
	Installments_Month int     `json:"installments_month"`
	Net_installment    float64 `json:"net_installment"`

	Total_Price      float64 `json:"total_price"`
	Paid_Amount      int     `json:"paid_amount"`
	Remaining_Amount float64 `json:"remaining_amount"`

	Total_Installments     int `json:"total_installments"`
	Paid_Installments      int `json:"paid_installments"`
	Remaining_Installments int `json:"remaining_installments"`

	Late_Day   int     `json:"late_day"`
	Fee_Amount float64 `json:"fee_amount"`

	Status          int     `json:"status"`
	Note            string  `json:"note"`
	TermType        int     `json:"term_type"`
	Interest_Amount float64 `json:"interest_amount"`
	Loan_Amount     float64 `json:"loan_amount"`

	Credit_Balance float64                     `json:"credit_balance"`
	Data           []Bill_Details_Installment2 `json:"data"` // 👈 ต้องมี

}
type BillService interface {
	CreateBill(request NewBillHeader) (*Bill_HeaderResponse, error)
	AddExtraPayment(billID uint, installmentID uint, request UpdateAddExtraRequest) error
	PayInstallment(billID uint, detailID uint, amount float64) ([]InstallmentPayResult, error)
	AutoApplyLateFees() error
	GetAllBill(
		invs []string,
		dateFrom, dateTo *time.Time,
		page, limit int,
		sortOrder int, // <-- เพิ่มตรงนี้
	) (*PaginationResponseBill, error)
	GetBillById(id uint) (*Bill_HeaderResponse, error)
	GetBillDetailById(id uint) (*Bill_DetailsResponse, error)
	CreateInstallmentBill(request NewInstallmentBillHeader, installMentId uint) (*Bill_HeaderResponse_Installment, error)
	GetInstallmentBillById(id uint) (*Bill_HeaderResponse_Installment, error)
	GetInstallmentBillDetailById(id uint) (*Bill_Details_Installment, error)
	PayPurchaseInstallment(billID uint, detailID uint, amount float64) ([]InstallmentPayResult, error)
	AutoApplyInstallementLateFees() error
	//  AutoApplyInstallmentLateFees() error
	AddInstallmentExtraPayment(billID uint, installmentID uint, request UpdateAddExtraRequest_Installment) error
	GetAllInstallmentBill(
		invs []string,
		dateFrom, dateTo *time.Time,
		page, limit int,
		sortOrder int, // <-- เพิ่มตรงนี้
	) (*PaginationResponseBillInstallment, error)
	UpdateBill(id uint, request Update_Installment) (*Bill_HeaderResponse, error)
	UpdateBill_Installment(id uint, request Update_Installment) (*Bill_HeaderResponse_Installment, error)
	GetAllBillUnpay(
		invs []string,
		dateFrom, dateTo *time.Time,
		page, limit int,
		sortOrder int, // <-- เพิ่มตรงนี้
		nameOrPhone []string,
	) (*PaginationResponseBill, error)
	GetAllInstallmentBillUnpay(
		invs []string,
		dateFrom, dateTo *time.Time,
		page, limit int,
		sortOrder int, // <-- เพิ่มตรงนี้
		nameOrPhone []string,

	) (*PaginationResponseBillInstallment, error)
	GetBillDetailsByIdUnpaid(id uint) ([]Bill_DetailsResponse, error)
	GetInstallmentBillByIdUnpaid(id uint) ([]Bill_Details_Installment, error)
	GetDueTodayBillsWithInstallments(sortData string) (*BillResponseWrapperMap, error)
	GetDueTodayInstallmentBillsWithInstallments(sortData string) (*BillResponseWrapperMapInstall, error)
	GetUnpaidBillById(userId string) ([]Bill_HeaderResponse1, error)
	GetUnpaidInstallmentBillById(userId string) ([]Bill_HeaderResponse_Installment1, error)
	GetpaidBillById(billID uint, detailID uint) ([]Bill_HeaderResponse1, error)
	GetpaidInstallmentBillById(billID uint, detailID uint) ([]Bill_HeaderResponse_Installment1, error)

	UpdateDailyInterest() error
	UpdateDailyInterestSingle(testDate ...time.Time) error
	RenewInterest(billID uint, payAmount float64, payDate time.Time ) (*model.Bill_Header_Installment, error)

	ApplyLateFeeToSingleBill(billID uint, today time.Time) error
		// UpdateDailyInterest1() error

}
