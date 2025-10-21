package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"rrmobile/service"
	"rrmobile/util"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ProductCategoryHandler interface {
	GetAllCategories(c *fiber.Ctx) error
	GetCategoryByID(c *fiber.Ctx) error
	CreateCategory(c *fiber.Ctx) error
	UpdateCategory(c *fiber.Ctx) error
	DeleteCategory(c *fiber.Ctx) error
}

type ProductRequestHandler interface {
	GetAllProducts(c *fiber.Ctx) error
	AddProduct(c *fiber.Ctx) error
	UpdateProduct(c *fiber.Ctx) error
	DeleteProduct(c *fiber.Ctx) error
	GetProductByID(c *fiber.Ctx) error
	GetProductByIDDetail(c *fiber.Ctx) error
}

type productCategoryHandler struct {
	productCategoryService service.ProductCategoryService
}

func NewProductCategoryHandler(productCategoryService service.ProductCategoryService) *productCategoryHandler {
	return &productCategoryHandler{productCategoryService: productCategoryService}
}

type productHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *productHandler {
	return &productHandler{productService: productService}
}
func (h *productCategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req service.NewProductCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	category, err := h.productCategoryService.CreateCategory(req)
	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "Failed to create category",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(category)
}
func (h *productCategoryHandler) GetAllCategories(c *fiber.Ctx) error {
	name := c.Query("name", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", ""))

	resp, err := h.productCategoryService.GetCategories(name, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(resp)
}
func (h *productCategoryHandler) GetCategoryByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}
	category, err := h.productCategoryService.GetCategoryByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch category",
		})
	}
	return c.JSON(category)
}

func (h *productCategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}
	var req service.UpdateProductCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	category, err := h.productCategoryService.EditCategory(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update category",
		})
	}
	return c.JSON(category)
}

func (h *productCategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}
	if err := h.productCategoryService.DeleteCategory(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete category",
		})
	}
	return c.Status(fiber.StatusNoContent).SendString("Category deleted successfully")
}

func (h *productHandler) AddProduct(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid form-data"})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No images uploaded"})
	}

	imgUrls := make([]string, 0, len(files))

	// ตรวจสอบไฟล์ทั้งหมดก่อน
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		mime := file.Header.Get("Content-Type")

		if !isValidImage(ext, mime) {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Invalid image file: %s", file.Filename)})
		}
	}

	// ถ้าไฟล์ถูกต้องทั้งหมด เริ่มบันทึกและ resize
	for _, file := range files {
		newName := util.GenerateFileName(file.Filename) // สร้างชื่อไฟล์ใหม่
		savePath := fmt.Sprintf("../uploads/%s", newName)

		// บันทึกไฟล์ต้นฉบับชั่วคราว
		tempPath := fmt.Sprintf("../uploads/temp_%s", newName)
		if err := c.SaveFile(file, tempPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
		}

		// Resize ภาพ
		if err := util.ResizeImage(tempPath, savePath, 800, 800); err != nil {
			os.Remove(tempPath)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to resize image"})
		}

		// ลบไฟล์ต้นฉบับชั่วคราว
		os.Remove(tempPath)

		imgUrls = append(imgUrls, "../uploads/"+newName)
	}

	categoryID, err := strconv.Atoi(c.FormValue("category_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid category ID"})
	}

	price := c.FormValue("price")
	if price == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Price is required"})
	}

	isActive, _ := strconv.ParseBool(c.FormValue("is_active"))

	req := service.NewProductRequest{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Price:       price,
		IsActive:    isActive,
		CategoryId:  uint(categoryID),
		Images:      imgUrls,
	}

	product, err := h.productService.CreateProduct(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(product)
}

// ตรวจสอบว่าเป็นไฟล์ภาพ
func isValidImage(ext, mime string) bool {
	validExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}

	validMime := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}

	return validExt[ext] && validMime[mime]
}
func (h *productHandler) GetAllProducts(c *fiber.Ctx) error {
	skus := strings.Split(c.Query("sku", ""), ",")
	names := strings.Split(c.Query("search", ""), ",")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", ""))

	// Optional filters
	var categoryID *uint
	if cid := c.Query("category_id", ""); cid != "" {
		if id, err := strconv.Atoi(cid); err == nil {
			tmp := uint(id)
			categoryID = &tmp
		}
	}

	var isActive *bool
	if val := c.Query("is_active"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			isActive = &b
		}
	}

	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}

	sortPrice := c.Query("sort_price", "") // "asc" / "desc"

	resp, err := h.productService.GetAllProducts(
		skus, names, categoryID, isActive, dateFrom, dateTo, sortPrice, page, limit,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}

func (h *productHandler) UpdateProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product ID"})
	}

	name := c.FormValue("name")
	description := c.FormValue("description")
	categoryId, _ := strconv.Atoi(c.FormValue("category_id"))
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
	isActive, _ := strconv.ParseBool(c.FormValue("is_active"))

	// parse delete image IDs
	deleteIDsStr := c.FormValue("delete_image_ids")
	deleteIDs := []uint{}
	if deleteIDsStr != "" {
		for _, v := range strings.Split(deleteIDsStr, ",") {
			if num, _ := strconv.Atoi(strings.TrimSpace(v)); num > 0 {
				deleteIDs = append(deleteIDs, uint(num))
			}
		}
	}



	uploadDir := "../uploads" // relative path กับ project
	os.MkdirAll(uploadDir, os.ModePerm)

	form, _ := c.MultipartForm()
	newImagesPaths := []string{}
	replaceImages := map[uint]string{}

	if form != nil {
		// Add new images
		if files, ok := form.File["new_images"]; ok {
			for _, f := range files {
				filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), f.Filename)
				relPath := fmt.Sprintf("%s/%s", uploadDir, filename) // ../uploads/xxx.jpg
				if err := c.SaveFile(f, relPath); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save file"})
				}
				newImagesPaths = append(newImagesPaths, relPath) // เก็บ ../uploads/xxx.jpg
			}
		}

		// Replace images
		for key, files := range form.File {
			if strings.HasPrefix(key, "replace_") && len(files) > 0 {
				imgIDStr := strings.TrimPrefix(key, "replace_")
				imgID, convErr := strconv.Atoi(imgIDStr)
				if convErr != nil || imgID == 0 {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid image ID for %s", key)})
				}

				filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), files[0].Filename)
				relPath := fmt.Sprintf("%s/%s", uploadDir, filename) // ../uploads/xxx.jpg
				if err := c.SaveFile(files[0], relPath); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save file"})
				}
				replaceImages[uint(imgID)] = relPath
			}
		}
	}


	req := service.UpdateProductRequest{
		Name:           name,
		Description:    description,
		CategoryId:     uint(categoryId),
		Price:          price,
		IsActive:       isActive,
		NewImagesPaths: newImagesPaths,
		ReplaceImages:  replaceImages,
		DeleteImageIDs: deleteIDs,
	}

	updated, err := h.productService.EditProduct(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(updated)
}

func (h *productHandler) DeleteProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product ID"})
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)

}

func (h *productHandler) GetProductByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	product, err := h.productService.GetProductByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(product)
}

// GetProductByIDDetail(productID uint) (*ProductResponse, error)
func (h *productHandler) GetProductByIDDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	product, err := h.productService.GetProductByIDDetail(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(product)
}
