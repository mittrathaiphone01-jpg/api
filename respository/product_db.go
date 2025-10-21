package respository

import (
	"errors"
	"fmt"
	"os"
	"rrmobile/model"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type productRepositoryDB struct {
	db *gorm.DB
}

func NewProductRepositoryDB(db *gorm.DB) ProductRepository {
	return &productRepositoryDB{db: db}
}

// productCategoryRepositoryDB implements ProductCategoryRepository
type productCategoryRepositoryDB struct {
	db *gorm.DB
}

func NewProductCategoryRepositoryDB(db *gorm.DB) ProductCategoryRepository {
	return &productCategoryRepositoryDB{db: db}
}

// productImageRepositoryDB implements ProductImageRepository
type productImageRepositoryDB struct {
	db *gorm.DB
}

func NewProductImageRepositoryDB(db *gorm.DB) ProductImageRepository {
	return &productImageRepositoryDB{db: db}
}
func (r *productCategoryRepositoryDB) GetCategories(names []string, limit, offset int) ([]ProductCategory, error) {
	var categories []ProductCategory
	query := r.db.Model(&ProductCategory{})

	if len(names) > 0 {
		// ใช้ OR + ILIKE สำหรับแต่ละค่า
		for i, n := range names {
			if i == 0 {
				query = query.Where("name ILIKE ?", "%"+n+"%")
			} else {
				query = query.Or("name ILIKE ?", "%"+n+"%")
			}
		}
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&categories).Error
	return categories, err
}

func (r *productCategoryRepositoryDB) CountCategories(names []string) (int64, error) {
	var count int64
	query := r.db.Model(&ProductCategory{})

	if len(names) > 0 {
		for i, n := range names {
			if i == 0 {
				query = query.Where("name ILIKE ?", "%"+n+"%")
			} else {
				query = query.Or("name ILIKE ?", "%"+n+"%")
			}
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *productCategoryRepositoryDB) GetCategoryByID(id uint) (*ProductCategory, error) {
	if id == 0 {
		return nil, errors.New("invalid category ID: ID cannot be zero")
	}
	var category ProductCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *productCategoryRepositoryDB) AddCategory(category ProductCategory) (*ProductCategory, error) {
	var count int64
	if err := r.db.Model(&ProductCategory{}).Count(&count).Error; err != nil {
		return nil, err
	}

	if count >= 5 {
		return nil, fmt.Errorf("ไม่สามารถเพิ่มหมวดหมู่ได้: มีหมวดหมู่สูงสุด 5 รายการแล้ว")
	}

	existingCategory := ProductCategory{}
	tx := r.db.Where("name = ?", category.Name).First(&existingCategory)
	if tx.RowsAffected > 0 {
		return nil, gorm.ErrDuplicatedKey
	}

	if err := r.db.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}


func (r *productCategoryRepositoryDB) UpdateCategory(id uint, category ProductCategory) (*ProductCategory, error) {
	if id == 0 {
		return nil, errors.New("invalid category ID: ID cannot be zero")
	}

	// เริ่ม transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// ตรวจสอบ duplicate name ก่อน update
	if category.Name != "" {
		var count int64
		if err := tx.Model(&ProductCategory{}).
			Where("name = ? AND id != ?", category.Name, id).
			Count(&count).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if count > 0 {
			tx.Rollback()
			return nil, fmt.Errorf("category with name  already exists", category.Name)
		}
	}

	// Update เฉพาะ field ที่มีค่า
	err := tx.Model(&ProductCategory{}).
		Where("id = ?", id).
		Updates(category).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ดึงข้อมูลใหม่หลัง update
	var updated ProductCategory
	if err := tx.First(&updated, id).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &updated, nil
}

func (r *productCategoryRepositoryDB) DeleteCategory(id uint) error {
	if id == 0 {
		return errors.New("invalid category ID: ID cannot be zero")
	}

	// delete แล้วเก็บผลลัพธ์
	tx := r.db.Where("id = ?", id).Delete(&ProductCategory{})

	if tx.Error != nil {
		return fmt.Errorf("failed to delete category with ID ", id, tx.Error)
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("no category found with ID ", id)
	}

	return nil
}

func (r *productRepositoryDB) GetLastSKUByYear(yearSuffix string) (string, error) {
	var lastSKU string
	err := r.db.Model(&Product{}).
		Select("sku").
		Where("sku LIKE ?", "A"+yearSuffix+"-%").
		Order("sku DESC").
		Limit(1).
		Scan(&lastSKU).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	return lastSKU, nil
}
func GenerateSKU(r ProductRepository) (string, error) {
	year := time.Now().Year() + 543
	yearSuffix := fmt.Sprintf("%02d", year%100)
	lastSKU, err := r.GetLastSKUByYear(yearSuffix)
	if err != nil {
		return "", err
	}

	var nextSeq int
	if lastSKU == "" {
		nextSeq = 1
	} else {
		parts := strings.Split(lastSKU, "-")
		lastNum, _ := strconv.Atoi(parts[1])
		nextSeq = lastNum + 1
	}

	return fmt.Sprintf("A%s-%03d", yearSuffix, nextSeq), nil
}

func (r *productRepositoryDB) AddProduct(product Product, images []ProductImage) (*Product, error) {
	var result Product
	savedFiles := []string{}
	for _, img := range images {
		savedFiles = append(savedFiles, img.ImageUrl)
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return fmt.Errorf("ไม่สามารถสร้างสินค้าได้: ชื่อสินค้าซ้ำ")
			}
			return fmt.Errorf("ไม่สามารถสร้างสินค้าได้: ", err)
		}

		for _, img := range images {
			img.ProductId = product.Id
			if err := tx.Create(&img).Error; err != nil {
				return fmt.Errorf("ไม่สามารถบันทึกรูปภาพสินค้าได้: ", err)

			}
		}

		if err := tx.Preload("ProductImages").First(&result, product.Id).Error; err != nil {
			return fmt.Errorf("ไม่สามารถโหลดข้อมูลสินค้าที่สร้างได้:", err)

		}
		return nil
	})

	if err != nil {
		// rollback ไฟล์ที่ save ไปแล้ว
		for _, f := range savedFiles {
			os.Remove("." + f)
		}
		return nil, err
	}

	return &result, nil
}

func (r *productRepositoryDB) GetAllProducts(filter ProductFilter, limit, offset int) ([]Product, error) {
	var products []Product
	query := r.db.Model(&Product{}).Preload("ProductImages").Preload("Category")

	// SKU filter
	var cleanSKUs []string
	for _, sku := range filter.SKUs {
		if sku != "" {
			cleanSKUs = append(cleanSKUs, sku)
		}
	}
	if len(cleanSKUs) > 0 {
		for i, sku := range cleanSKUs {
			pattern := "%" + sku + "%"
			if i == 0 {
				query = query.Where("sku ILIKE ?", pattern)
			} else {
				query = query.Or("sku ILIKE ?", pattern)
			}
		}
	}

	// Name filter
	var likeNames []string
	for _, name := range filter.Names {
		if name != "" {
			likeNames = append(likeNames, "%"+name+"%")
		}
	}
	if len(likeNames) > 0 {
		query = query.Where("name ILIKE ANY(?)", pq.Array(likeNames))
	}

	// Category / isActive / Date filters
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	// Sorting
	switch filter.SortPrice {
	case "asc":
		query = query.Order("price ASC")
	case "desc":
		query = query.Order("price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// ✅ Apply limit/offset only if limit > 0
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepositoryDB) CountProducts(filter ProductFilter) (int64, error) {
	var count int64
	query := r.db.Model(&Product{})

	// กรอง SKU ด้วย ILIKE OR เหมือน GetAllProducts
	if len(filter.SKUs) > 0 {
		var cleanSKUs []string
		for _, sku := range filter.SKUs {
			if sku != "" {
				cleanSKUs = append(cleanSKUs, sku)
			}
		}
		if len(cleanSKUs) > 0 {
			for i, sku := range cleanSKUs {
				pattern := "%" + sku + "%"
				if i == 0 {
					query = query.Where("sku ILIKE ?", pattern)
				} else {
					query = query.Or("sku ILIKE ?", pattern)
				}
			}
		}
	}

	// กรองชื่อเหมือน GetAllProducts
	if len(filter.Names) > 0 {
		var likeNames []string
		for _, name := range filter.Names {
			if name != "" {
				likeNames = append(likeNames, "%"+name+"%")
			}
		}
		if len(likeNames) > 0 {
			query = query.Where("name ILIKE ANY(?)", pq.Array(likeNames))
		}
	}

	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}


func (r *productRepositoryDB) UpdateProductWithImages(
	id uint,
	product Product,
	newImages []string,
	replaceImages map[uint]string,
	deleteImageIDs []uint,
) (*Product, error) {
	if id == 0 {
		return nil, errors.New("invalid product ID")
	}

	tx := r.db.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()

	savedFiles := []string{}

	// Load existing product
	var existingProduct Product
	if err := tx.First(&existingProduct, id).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update Product fields
	updates := map[string]interface{}{}
	if product.Name != "" {
		updates["name"] = product.Name
	}
	if product.Sku != "" {
		updates["sku"] = product.Sku
	}
	if product.Description != "" {
		updates["description"] = product.Description
	}
	if product.Price != 0 {
		updates["price"] = product.Price
	}
	updates["is_active"] = product.IsActive
	if product.CategoryId != 0 {
		updates["category_id"] = product.CategoryId
	}

	if err := tx.Model(&Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete images
	if len(deleteImageIDs) > 0 {
		var oldImages []ProductImage
		if err := tx.Where("id IN ? AND product_id = ?", deleteImageIDs, id).Find(&oldImages).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, img := range oldImages {
			if _, err := os.Stat(img.ImageUrl); err == nil {
				os.Remove(img.ImageUrl)
			}
		}
		if err := tx.Where("id IN ? AND product_id = ?", deleteImageIDs, id).Delete(&ProductImage{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Replace images
	for imgID, newRelPath := range replaceImages {
		var oldImg ProductImage
		if err := tx.First(&oldImg, imgID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("cannot replace image: image ID  not found", imgID)
		}

		// ใช้ relative path เดิม + move ไฟล์
		if err := os.Rename(newRelPath, oldImg.ImageUrl); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to replace image: ", err)
		}
		// DB ไม่ต้อง update เพราะ relative path เดิมอยู่แล้ว
	}

	// Add new images
	for _, path := range newImages {
		newImg := ProductImage{
			ProductId: id,
			ImageUrl:  path,
		}
		if err := tx.Create(&newImg).Error; err != nil {
			// rollback newly saved files
			for _, f := range savedFiles {
				os.Remove(f)
			}
			tx.Rollback()
			return nil, err
		}
		savedFiles = append(savedFiles, path)
	}

	// Preload updated product
	var updated Product
	if err := tx.Preload("Category").Preload("ProductImages").First(&updated, id).Error; err != nil {
		for _, f := range savedFiles {
			os.Remove(f)
		}
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &updated, nil
}

func (r *productRepositoryDB) DeleteProductImages(tx *gorm.DB, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Where("id IN ?", ids).Delete(&model.ProductImage{}).Error
}

func (r *productRepositoryDB) DeleteProductWithImages(id uint) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var product Product
	if err := tx.Preload("ProductImages").First(&product, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// ลบ record ใน DB ก่อน
	if err := tx.Where("product_id = ?", id).Delete(&ProductImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&Product{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// ลบไฟล์ใน Disk แบบไม่ error ถ้าไฟล์ไม่เจอ
	for _, img := range product.ProductImages {
		if _, err := os.Stat(img.ImageUrl); err == nil {
			_ = os.Remove(img.ImageUrl)
		}
	}

	tx.Commit()
	return nil
}
func (r *productRepositoryDB) GetProductByID(id uint) (*Product, error) {
	var product Product
	if err := r.db.Preload("ProductImages").Preload("Category").First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
