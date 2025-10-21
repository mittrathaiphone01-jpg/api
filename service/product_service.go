package service

import (
	"fmt"
	"path/filepath"
	"rrmobile/respository"
	"rrmobile/util"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type productCategoryService struct {
	productcategoryRepository respository.ProductCategoryRepository
}

func NewProductCategoryService(productcategoryRepository respository.ProductCategoryRepository) ProductCategoryService {
	return &productCategoryService{productcategoryRepository: productcategoryRepository}
}

type productService struct {
	db *gorm.DB

	productRepository respository.ProductRepository
}

func NewProductService(productRepository respository.ProductRepository) *productService {
	return &productService{productRepository: productRepository}
}
func (s *productCategoryService) GetCategories(name string, page, limit int) (*PaginationResponseCategory, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	names := strings.Split(name, ",")
	if len(names) == 1 && names[0] == "" {
		names = nil // ถ้าไม่มีชื่อให้ค้นหา ให้ใช้ nil แทน
	}
	total, err := s.productcategoryRepository.CountCategories(names)
	if err != nil {
		return nil, fmt.Errorf("failed to count categories: ", err)
	}

	categories, err := s.productcategoryRepository.GetCategories(names, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories:", err)
	}

	var categoryResponses []ProductCategoryResponse
	for _, c := range categories {
		categoryResponses = append(categoryResponses, ProductCategoryResponse{
			Id:        c.Id,
			Name:      c.Name,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
			IsActive:  c.IsActive,
		})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginationResponseCategory{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		Limit:       limit,
		Categories:  categoryResponses,
	}, nil
}
func (s *productCategoryService) GetCategoryByID(id uint) (*ProductCategoryResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category ID")
	}
	category, err := s.productcategoryRepository.GetCategoryByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category by ID: ", err)
	}
	return &ProductCategoryResponse{
		Id:        category.Id,
		Name:      category.Name,
		CreatedAt: category.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: category.UpdatedAt.Format("2006-01-02 15:04:05"),
		IsActive:  category.IsActive,
	}, nil
}
func (s *productCategoryService) CreateCategory(req NewProductCategoryRequest) (*ProductCategoryResponse, error) {
	// แปลง request → entity
	category := respository.ProductCategory{
		Name:     req.Name,
		IsActive: true, // กำหนดค่าเริ่มต้นเป็น true
	}

	// เรียกใช้ Repository
	createdCategory, err := s.productcategoryRepository.AddCategory(category)
	if err != nil {
		return nil, err
	}

	// แปลง entity → response
	return &ProductCategoryResponse{
		Id:        createdCategory.Id,
		Name:      createdCategory.Name,
		CreatedAt: createdCategory.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: createdCategory.UpdatedAt.Format("2006-01-02 15:04:05"),
		IsActive:  createdCategory.IsActive,
	}, nil
}

func (s *productCategoryService) EditCategory(id uint, updateCategoryRequest UpdateProductCategoryRequest) (*ProductCategoryResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category ID")
	}
	category := respository.ProductCategory{
		Name:     updateCategoryRequest.Name,
		IsActive: updateCategoryRequest.IsActive,
	}
	updatedCategory, err := s.productcategoryRepository.UpdateCategory(id, category)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: ", err)
	}
	return &ProductCategoryResponse{
		Id:        updatedCategory.Id,
		Name:      updatedCategory.Name,
		CreatedAt: updatedCategory.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: updatedCategory.UpdatedAt.Format("2006-01-02 15:04:05"),
		IsActive:  updatedCategory.IsActive,
	}, nil
}

func (s *productCategoryService) DeleteCategory(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid category ID")
	}
	err := s.productcategoryRepository.DeleteCategory(id)
	if err != nil {
		return fmt.Errorf("failed to delete category: ", err)
	}
	return nil
}

func (s *productService) CreateProduct(req NewProductRequest) (*ProductResponse, error) {
	// แปลงราคา
	priceFloat, err := strconv.ParseFloat(req.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price: ", err)
	}

	// Generate SKU
	sku, err := respository.GenerateSKU(s.productRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SKU: ", err)
	}

	// สร้าง product entity
	product := respository.Product{
		Sku:         sku,
		Name:        req.Name,
		Description: req.Description,
		Price:       priceFloat,
		IsActive:    true,
		CategoryId:  req.CategoryId,
	}

	// สร้าง images entities สำหรับ DB
	var imagesEntities []respository.ProductImage
	for _, url := range req.Images {
		imagesEntities = append(imagesEntities, respository.ProductImage{
			ImageUrl: url,
		})
	}

	// เพิ่มลง DB
	createdProduct, err := s.productRepository.AddProduct(product, imagesEntities)
	if err != nil {
		return nil, err
	}

	// แปลง images สำหรับ response
	var imagesResponse []ProductImageResponse
	for _, img := range createdProduct.ProductImages {
		imagesResponse = append(imagesResponse, ProductImageResponse{
			Id:  img.Id,
			Url: img.ImageUrl,
		})
	}

	return &ProductResponse{
		Id:          createdProduct.Id,
		Name:        createdProduct.Name,
		Sku:         createdProduct.Sku,
		Description: createdProduct.Description,
		Price:       fmt.Sprintf("%.2f", createdProduct.Price),
		Category:    createdProduct.Category.Name,
		IsActive:    createdProduct.IsActive,
		Images:      imagesResponse,
		CreatedAt:   createdProduct.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   createdProduct.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
func (s *productService) GenerateProductResponse(products []ProductResponse) []ProductResponse {
	resp := make([]ProductResponse, 0, len(products))

	for _, p := range products {
		imagesResponse := []ProductImageResponse{}

		for _, img := range p.Images {
			filename := img.Url // จะเป็นชื่อไฟล์จริง
			token, _ := util.GenerateImageToken(filename)
			imagesResponse = append(imagesResponse, ProductImageResponse{
				Id:  img.Id,
				Url: fmt.Sprintf("/image?token=%s", token),
			})
		}

		resp = append(resp, ProductResponse{
			Id:        p.Id,
			Sku:       p.Sku,
			Name:      p.Name,
			Images:    imagesResponse,
			Price:     fmt.Sprintf("%.2f", p.Price),
			Category:  p.Category, // เป็น string
			IsActive:  p.IsActive,
			CreatedAt: p.CreatedAt, // เป็น string จาก DB
			UpdatedAt: p.UpdatedAt,
		})
	}

	return resp
}

func (s *productService) GetAllProducts(
	skus, names []string,
	categoryID *uint,
	isActive *bool,
	dateFrom, dateTo *time.Time,
	sortPrice string,
	page, limit int,
) (*PaginationResponseProduct, error) {

	if page < 1 {
		page = 1
	}

	offset := 0
	if limit > 0 {
		offset = (page - 1) * limit
	}

	filter := respository.ProductFilter{
		SKUs:       skus,
		Names:      names,
		CategoryID: categoryID,
		IsActive:   isActive,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		SortPrice:  sortPrice,
	}

	// ดึงจำนวนทั้งหมด
	total, err := s.productRepository.CountProducts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: ", err)
	}

	// ดึงรายการสินค้า (limit=0 → ไม่มีการจำกัด)
	products, err := s.productRepository.GetAllProducts(filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: ", err)
	}

	// สร้าง product responses
	var productResponses []ProductResponse
	for _, p := range products {
		var imagesResponse []ProductImageResponse
		for _, img := range p.ProductImages {
			filename := filepath.Base(img.ImageUrl)
			token, _ := util.GenerateImageToken(filename)
			imagesResponse = append(imagesResponse, ProductImageResponse{
				Id:  img.Id,
				Url: fmt.Sprintf("/image?token=%s", token),
			})
		}

		categoryName := ""
		if p.Category.Name != "" {
			categoryName = p.Category.Name
		}

		productResponses = append(productResponses, ProductResponse{
			Id:          p.Id,
			Sku:         p.Sku,
			Name:        p.Name,
			Description: p.Description,
			Price:       fmt.Sprintf("%.2f", p.Price),
			Category:    categoryName,
			CategoryId:  p.CategoryId,
			IsActive:    p.IsActive,
			Images:      imagesResponse,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// pagination เฉพาะเมื่อ limit > 0
	var totalPages int
	var hasNext, hasPrev bool

	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
		hasNext = page < totalPages
		hasPrev = page > 1
	} else {
		// แสดงทั้งหมด
		totalPages = 1
		hasNext = false
		hasPrev = false
	}

	return &PaginationResponseProduct{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		Limit:       limit,
		Products:    productResponses,
	}, nil
}

func (s *productService) EditProduct(id uint, req UpdateProductRequest) (*ProductResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid product ID")
	}

	product := respository.Product{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Price:       req.Price,
	}

	// ตรวจสอบ CategoryId
	if req.CategoryId != 0 {
		product.CategoryId = req.CategoryId
	}

	updatedProduct, err := s.productRepository.UpdateProductWithImages(
		id,
		product,
		req.NewImagesPaths,
		req.ReplaceImages,
		req.DeleteImageIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: ", err)
	}

	var imagesResponse []ProductImageResponse
	for _, img := range updatedProduct.ProductImages {
		filename := filepath.Base(img.ImageUrl)
		token, _ := util.GenerateImageToken(filename)
		imagesResponse = append(imagesResponse, ProductImageResponse{
			Id:  img.Id,
			Url: fmt.Sprintf("/image?token=%s", token),
		})
	}
	categoryName := ""
	if updatedProduct.Category.Id != 0 {
		categoryName = updatedProduct.Category.Name
	}

	return &ProductResponse{
		Id:          updatedProduct.Id,
		Name:        updatedProduct.Name,
		Sku:         updatedProduct.Sku,
		Description: updatedProduct.Description,
		Price:       fmt.Sprintf("%.2f", updatedProduct.Price),
		Category:    categoryName,
		IsActive:    updatedProduct.IsActive,
		Images:      imagesResponse, // <- ต้องเป็น []ProductImageResponse
		CreatedAt:   updatedProduct.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   updatedProduct.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *productService) DeleteProduct(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid product ID")
	}

	return s.productRepository.DeleteProductWithImages(id)
}

func (s *productService) GetProductByID(productID uint) (*ProductResponse, error) {
	product, err := s.productRepository.GetProductByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: ", err)
	}

	var imagesResponse []ProductImageResponse
	for _, img := range product.ProductImages {
		imagesResponse = append(imagesResponse, ProductImageResponse{
			Id:  img.Id,
			Url: img.ImageUrl,
		})
	}
	resp := &ProductResponse{
		Id:          product.Id,
		Sku:         product.Sku,
		Name:        product.Name,
		Description: product.Description,
		Price:       fmt.Sprintf("%.2f", product.Price),
		Category:    product.Category.Name, // เป็น string
		CategoryId:  product.Category.Id,   // เป็น string
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
		Images:      imagesResponse,
	}

	return resp, nil
}

func (s *productService) GetProductByIDDetail(productID uint) (*ProductResponse, error) {
	product, err := s.productRepository.GetProductByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// สร้าง images พร้อม token 1 นาที
	imagesResponse := []ProductImageResponse{}
	for _, img := range product.ProductImages {
		filename := img.ImageUrl                      // ชื่อไฟล์จริงจาก DB
		token, _ := util.GenerateImageToken(filename) // token 1 นาที
		imagesResponse = append(imagesResponse, ProductImageResponse{
			Id:  img.Id,
			Url: fmt.Sprintf("/image?token=%s", token),
		})
	}

	resp := &ProductResponse{
		Id:          product.Id,
		Sku:         product.Sku,
		Name:        product.Name,
		Description: product.Description,
		Price:       fmt.Sprintf("%.2f", product.Price),
		Category:    product.Category.Name, // เป็น string
		CategoryId:  product.Category.Id,   // เป็น string
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
		Images:      imagesResponse,
	}

	return resp, nil
}
