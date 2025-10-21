package service

import (
	"time"
)

type ProductResponseAll struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Sku         string `json:"sku"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Category    string `json:"category"`
	IsActive    bool   `json:"is_active"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	ItemsLeft   int    `json:"items_left"`
}
type PaginatedProductResponse struct {
	Products    []ProductResponse `json:"data"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	ItemsLeft   int               `json:"items_left"`
}

type NewProductRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Price       string   `json:"price" validate:"required,number"`
	CategoryId  uint     `json:"category_id" validate:"required"`
	IsActive    bool     `json:"is_active"`
	Images      []string `json:"images"` // list ของ URL รูปภาพ
}
type ProductImageResponse struct {
	Id  uint   `json:"id"`
	Url string `json:"url"`
}
type ProductResponse struct {
	Id          uint   `json:"id"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Category    string `json:"category"`
	CategoryId  uint   `json:"category_id"`

	IsActive  bool                   `json:"is_active"`
	Images    []ProductImageResponse `json:"images"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type UpdateProductRequest struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	CategoryId     uint            `json:"category_id"`
	Price          float64         `json:"price"`
	IsActive       bool            `json:"is_active"`
	NewImagesPaths []string        `json:"-"` // path ของรูปใหม่หลัง save
	ReplaceImages  map[uint]string `json:"-"` // key=ID รูปเก่า, value= path ใหม่
	DeleteImageIDs []uint          `json:"delete_image_ids"`
}

type ProductCategoryResponseAll struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	IsActive    bool   `json:"is_active"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	ItemsLeft   int    `json:"items_left"`
}
type PaginatedProductCategoryResponse struct {
	Categories  []ProductCategoryResponse `json:"results"`
	CurrentPage int                       `json:"current_page"`
	TotalPages  int                       `json:"total_pages"`
	ItemsLeft   int                       `json:"items_left"`
}
type ProductCategoryResponse struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsActive  bool   `json:"is_active"`
}

type PaginationResponseCategory struct {
	Total       int64                     `json:"total"`
	TotalPages  int                       `json:"total_pages"`
	CurrentPage int                       `json:"current_page"`
	HasNext     bool                      `json:"has_next"`
	HasPrev     bool                      `json:"has_prev"`
	Limit       int                       `json:"limit"`
	Categories  []ProductCategoryResponse `json:"categories"`
}
type PaginationResponseProduct struct {
	Total       int64             `json:"total"`
	TotalPages  int               `json:"total_pages"`
	CurrentPage int               `json:"current_page"`
	HasNext     bool              `json:"has_next"`
	HasPrev     bool              `json:"has_prev"`
	Limit       int               `json:"limit"`
	Products    []ProductResponse `json:"data"`
}
type NewProductCategoryRequest struct {
	Name     string `json:"name" validate:"required,unique"`
	IsActive bool   `json:"is_active"`
}
type UpdateProductCategoryRequest struct {
	Name     string `json:"name" validate:"required,unique"`
	IsActive bool   `json:"is_active"`
}

type ProductCategoryService interface {
	// CountCategories(name []string) (int64, error)
	GetCategories(name string, page, limit int) (*PaginationResponseCategory, error) // GetCategoryById(id uint) (*ProductCategoryResponse, error)
	CreateCategory(NewProductCategoryRequest) (*ProductCategoryResponse, error)
	EditCategory(id uint, updateCategoryRequest UpdateProductCategoryRequest) (*ProductCategoryResponse, error)
	DeleteCategory(id uint) error
	GetCategoryByID(id uint) (*ProductCategoryResponse, error)
}

type ProductService interface {
	GetAllProducts(
		skus, names []string,
		categoryID *uint,
		isActive *bool,
		dateFrom, dateTo *time.Time,
		sortPrice string,
		page, limit int,
	) (*PaginationResponseProduct, error)
	CreateProduct(req NewProductRequest) (*ProductResponse, error)
	EditProduct(id uint, req UpdateProductRequest) (*ProductResponse, error)
	DeleteProduct(id uint) error
	GetProductByID(productID uint) (*ProductResponse, error)
	GetProductByIDDetail(productID uint) (*ProductResponse, error)
}
