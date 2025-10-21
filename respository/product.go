package respository

import (
	"time"

	"gorm.io/gorm"
)

type ProductCategory struct {
	Id        uint      `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsActive  bool      `db:"is_active"`
}

type Product struct {
	Id            uint            `db:"id"`
	Sku           string          `db:"sku"`
	Name          string          `db:"name"`
	Description   string          `db:"description"`
	Price         float64         `db:"price"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	Category      ProductCategory `db:"category"`
	CategoryId    uint            `db:"category_id"`
	IsActive      bool            `db:"is_active"`
	ProductImages []ProductImage  `db:"product_images"`
}
type ProductImage struct {
	Id        uint      `db:"id"`
	ProductId uint      `db:"product_id"`
	ImageUrl  string    `db:"image_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
type ProductFilter struct {
	SKUs       []string
	
	Names      []string
	CategoryID *uint
	IsActive   *bool
	DateFrom   *time.Time
	DateTo     *time.Time
	SortPrice  string // "asc" หรือ "desc"
}
type ProductCategoryRepository interface {
	GetCategories(names []string, limit, offset int) ([]ProductCategory, error)

	AddCategory(category ProductCategory) (*ProductCategory, error)
	UpdateCategory(id uint, category ProductCategory) (*ProductCategory, error)
	DeleteCategory(id uint) error
	CountCategories(names []string) (int64, error)
	GetCategoryByID(id uint) (*ProductCategory, error)

}
type ProductRepository interface {
	GetAllProducts(filter ProductFilter, limit, offset int) ([]Product, error)
	CountProducts(filter ProductFilter) (int64, error)
	GetProductByID(id uint) (*Product, error)
	GetLastSKUByYear(yearSuffix string) (string, error)
	AddProduct(product Product, images []ProductImage) (*Product, error)
	UpdateProductWithImages(id uint, product Product, newImages []string, replaceImages map[uint]string, deleteImageIDs []uint) (*Product, error)
	DeleteProductImages(tx *gorm.DB, ids []uint) error
	DeleteProductWithImages(id uint) error
}
type ProductImageRepository interface {

}
