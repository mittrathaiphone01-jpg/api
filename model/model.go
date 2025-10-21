package model

import (
	"time"

	"gorm.io/gorm"
)

type Users struct {
	Id         uint
	Username   string `gorm:"size:50;uniqueIndex;index:idx_username"`
	Password   string
	FullName   string `gorm:"size:50;uniqueIndex"`
	RoleID     int
	Role       Role      `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	Is_active  bool      `gorm:"default:true;index:idx_is_active"`
	Created_At time.Time `gorm:"autoCreateTime"`
	Updated_At time.Time `gorm:"autoUpdateTime"`
}
type RefreshToken struct {
	Id        uint      `gorm:"primaryKey"`
	User_Id   int       `gorm:"index"`
	User      Users     `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	TokenHash string    `gorm:"not null;index:idx_token_hash"`
	ExpiresAt time.Time `gorm:"index:expires_at"` // ใช้ time.Time แทน int เพื่อความชัดเจน
	CreatedAt time.Time
	IsRevoked bool `gorm:"default:false;index:idx_token_isrevoked"` // ใช้ bool แทน int เพื่อความชัดเจน
}
type AccessToken struct {
	Id        uint `gorm:"primaryKey"`
	User_Id   uint `gorm:"index"`
	Token     string
	ExpiresAt time.Time `gorm:"index"`
	IsRevoked bool      `gorm:"default:false;index"`
	CreatedAt time.Time
}

type Role struct {
	Id        uint `gorm:"primaryKey"`
	Role_Name string
}

type ProductCategory struct {
	Id        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;uniqueIndex;index:idx_product_category_name"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsActive  bool      `gorm:"default:true;index:idx_product_category_is_active"`
}

type Product struct {
	Id            uint            `gorm:"primaryKey"`
	Sku           string          `gorm:"size:50;uniqueIndex;index:idx_product_sku"`
	Name          string          `gorm:"size:50;uniqueIndex;index:idx_product_name"`
	Description   string          `gorm:"size:50;index:idx_product_description"`
	Price         float64         `gorm:"type:decimal(10,2)"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	Category      ProductCategory `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	CategoryId    uint            `gorm:"index:idx_product_category_id"`
	IsActive      bool            `gorm:"default:true;index:idx_product_is_active"`
	ProductImages []ProductImage  `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
}

type ProductImage struct {
	Id        uint      `gorm:"primaryKey"`
	ProductId uint      `gorm:"index:idx_product_image_product_id"`
	ImageUrl  string    `gorm:"size:255;not null;index:idx_product_image_url"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Product   Product   `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
}
type Rules struct {
	Id               uint      `gorm:"primaryKey"`
	Threshold_Months int       `gorm:"index:idx_rules_threshold_months"`
	Type_Discount    bool      `gorm:"default:false;index:idx_rules_type_discount"`
	Discount_Amount  float64   `gorm:"type:decimal(10,2);index:idx_rules_discount_amount"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

type Installment struct {
	Id  uint `gorm:"primaryKey"`
	Day int  `gorm:"index:idx_installment_day"`
}
type Fine_System_Category struct {
	Id uint `gorm:"primaryKey"`

	Name string `gorm:"size:100;uniqueIndex;index:idx_finesysystem,_category_name"`
}
type Fine_System struct {
	Id                     uint                 `gorm:"primaryKey"`
	FineAmount             float64              `gorm:"type:decimal(10,2);index:idx_fine_system_fine"`
	Fine_System_Category   Fine_System_Category `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	Fine_System_CategoryId uint                 `gorm:"index:idx_finesysystem,_category_name"`
}

type Member struct {
	Id       uint   `gorm:"primaryKey"`
	FullName string `gorm:"size:255;not null;index:idx_fullname"`
	Tel      string `gorm:"size:10;not null;"`
	UserId   string `gorm:"size:100;index:idx_user_id"`
}

type Bill_Header struct {
	Id        uint `gorm:"primaryKey"`
	Invoice   string
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt time.Time `gorm:"autoCreateTime"`
	// Member    Member    `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	// MemberId  uint      `gorm:"index:idx_member_id"`
	MemberId uint   `gorm:"index:idx_member_id1"`
	Member   Member `gorm:"foreignKey:MemberId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	User_Id   int     `gorm:"index:idx_user_id1"`
	User      Users   `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	ProductId uint    `gorm:"index:idx_product_image_product_id1"`
	Product   Product `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`

	Extra_Percent      int
	Down_Percent       int
	Installments_Month int
	Net_installment    float64

	Total_Price      float64
	Paid_Amount      int
	Remaining_Amount float64

	Total_Installments     int
	Paid_Installments      int
	Remaining_Installments int

	Late_Day       int
	Fee_Amount     float64
	Credit_Balance float64 `gorm:"default:0"`

	Status int

	Note string `gorm:"type:text"`

	// 👇 ต้องเพิ่มตรงนี้
	BillDetails []Bill_Details `gorm:"foreignKey:Bill_HeaderId;references:Id"`
}

type Bill_Details struct {
	Id uint `gorm:"primaryKey"`
	// // Bill_Header       Bill_Header `gorm:"constraint:OnUpdate:CASCADE;OnDelete:RESTRICT;"`
	// Bill_HeaderId     uint
	BillHeader    Bill_Header `gorm:"foreignKey:Bill_HeaderId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Bill_HeaderId uint        `gorm:"index:idx_bill_header_id"`

	Installment_Price float64
	Paid_Amount       float64

	Payment_Date time.Time `gorm:"index_payment_date1"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	Fee_Amount float64
	Status     int

	Credit_Balance float64 `gorm:"default:0"`
	Payment_No     string
}

type Bill_Header_Installment struct {
	Id        uint      `gorm:"primaryKey"`
	Invoice   string    `gorm:"index_invoice"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt

	MemberId uint   `gorm:"index:idx_member_id2"`
	Member   Member `gorm:"foreignKey:MemberId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	User_Id int   `gorm:"index:idx_user_id2"`
	User    Users `gorm:"foreignKey:User_Id;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ProductId uint    `gorm:"index:idx_product_id2"`
	Product   Product `gorm:"foreignKey:ProductId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Installment_Day int
	Extra_Percent   int
	Net_installment float64

	Total_Price      float64
	Paid_Amount      int
	Remaining_Amount float64

	Total_Installments     int
	Paid_Installments      int
	Remaining_Installments int

	Late_Day       int
	Fee_Amount     float64
	Credit_Balance float64 `gorm:"default:0"`

	Status int `gorm:"index:idx_status"`

	Note string `gorm:"type:text"`

	TermType              int
	TermValue             int
	Loan_Amount           float64 // ยอดที่ลูกค้าขอ (ยอดหลัก)
	Interest_Amount       float64 // ดอกเบี้ยรวมที่คำนวณ (Principal * Extra_Percent)
	Total_Interest_Amount float64 // ดอกเบี้ยรวมที่คำนวณ (Principal * Extra_Percent)

	LastRenewDate time.Time
	NextDueDate   time.Time
	RenewCount    int

	// ✅ One-To-Many Relation
	BillDetailsInstallment []Bill_Details_Installment `gorm:"foreignKey:Bill_Header_InstallmentId"`
}

type Bill_Details_Installment struct {
	Id uint `gorm:"primaryKey"`

	Bill_Header_InstallmentId uint                    `gorm:"index:idx_bill_header_installment_id3"`
	Bill_Header_Installment   Bill_Header_Installment `gorm:"foreignKey:Bill_Header_InstallmentId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Installment_Price float64
	Paid_Amount       float64

	Payment_Date time.Time `gorm:"index_payment_date3"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Fee_Amount float64
	Status     int `gorm:"index:idx_status"`

	Credit_Balance float64 `gorm:"default:0"`
	Payment_No     string

	Is_Interest_Only bool
}
