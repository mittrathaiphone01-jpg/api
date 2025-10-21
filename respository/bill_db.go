package respository

import (
	"errors"
	"fmt"
	"rrmobile/model"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type billRepositoryDB struct {
	db *gorm.DB
}

func NewBillRepositoryDB(db *gorm.DB) BillRepository {
	return &billRepositoryDB{db: db}
}

func (r *billRepositoryDB) GetLastInvByYear(yearSuffix string) (string, error) {
	var lastSKU string
	err := r.db.Model(&Bill_Header{}).
		Select("invoice").
		// หาเฉพาะปี (ไม่สนใจเดือน)
		Where("invoice LIKE ?", fmt.Sprintf("BI/%%-%s-%%", yearSuffix)).
		Order("invoice DESC").
		Limit(1).
		Scan(&lastSKU).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	return lastSKU, nil
}

func GenerateInv(r BillRepository) (string, error) {
	now := time.Now()
	month := fmt.Sprintf("%02d", int(now.Month()))
	year := now.Year() + 543
	yearSuffix := fmt.Sprintf("%02d", year%100)

	lastSKU, err := r.GetLastInvByYear(yearSuffix)
	if err != nil {
		return "", err
	}

	nextSeq := 1
	if lastSKU != "" {
		// ตัวอย่าง lastSKU = "BI/09-68-0001"
		parts := strings.Split(lastSKU, "-")
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid SKU format: %s", lastSKU)
		}
		num, err := strconv.Atoi(parts[2]) // "0001"
		if err != nil {
			return "", fmt.Errorf("invalid sequence number in SKU: %s", lastSKU)
		}
		nextSeq = num + 1
	}

	// คืนค่าในรูปแบบ BI/09-68-0001
	return fmt.Sprintf("BI/%s-%s-%04d", month, yearSuffix, nextSeq), nil
}

func (r *billRepositoryDB) CreateBill(bill *Bill_Header) (*Bill_Header, error) {
	existingBillHeader := Bill_Header{}
	tx := r.db.Where("id = ? OR invoice = ?", bill.Id, bill.Invoice).First(&existingBillHeader)
	if tx.RowsAffected > 0 {
		return nil, gorm.ErrDuplicatedKey
	}
	if err := r.db.Create(&bill).Error; err != nil {
		return nil, err
	}
	return bill, nil
}
func (r *billRepositoryDB) CreateBillDetails(details []Bill_Details) error {
	if len(details) == 0 {
		return errors.New("no bill details provided")
	}

	// ตรวจสอบว่า record แรกมีอยู่แล้วหรือไม่ (ป้องกัน duplicated key)
	var existing Bill_Details
	tx := r.db.Where("id = ? AND bill_header_id = ?", details[0].Id, details[0].Bill_HeaderId).First(&existing)
	if tx.RowsAffected > 0 {
		return gorm.ErrDuplicatedKey
	}

	// Insert slice ทั้งหมด
	if err := r.db.Create(&details).Error; err != nil {
		return err
	}
	return nil
}

func (r *billRepositoryDB) GetBillById(id uint) (*Bill_Header, error) {
	var bill Bill_Header
	err := r.db.Preload("BillDetails").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User").
		First(&bill, id).Error
	return &bill, err
}
func (r *billRepositoryDB) GetBillByIdLate(id uint) (*Bill_Header, error) {
	var bill Bill_Header
	err := r.db.Preload("BillDetails").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User").
		First(&bill, id).Error
	return &bill, err
}
func (r *billRepositoryDB) GetBillDetailById(id uint) (*Bill_Details, error) {
	var details Bill_Details
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_id = ? AND status = 0", id).
		Order("payment_date ASC ").
		Find(&details).Error
	return &details, err
}

func (r *billRepositoryDB) GetUnpaidInstallments(billID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_id = ? AND status = ?", billID, 0).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidInstallments2(billID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_id = ?  ", billID).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}
func (r *billRepositoryDB) GetBillHeaderById(id uint) (*Bill_Header, error) {
	var header Bill_Header
	err := r.db.
		Preload("Member").
		Preload("Product").
		Preload("User").
		Where("id = ?", id).
		First(&header).Error

	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (r *billRepositoryDB) GetPaidInstallments(billID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		Where("bill_header_id = ? AND status = 1", billID).
		Order("payment_date ASC").
		Find(&details).Error
	return details, err
}

func (r *billRepositoryDB) UpdateBillDetail(installments []Bill_Details) error {
	// แก้ไข: loop แบบ index เพื่อให้เราได้ pointer ไปที่ element ใน slice จริงๆ
	for i := range installments {
		if err := r.db.Save(&installments[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *billRepositoryDB) UpdateSingleBillDetail(detail *Bill_Details) error {
	return r.db.Save(detail).Error
}

func (r *billRepositoryDB) UpdateBill(bill *Bill_Header) error {
	var existingBill Bill_Header

	// 1. ตรวจสอบว่าบิลที่ต้องการแก้ไขมีอยู่หรือไม่
	if err := r.db.First(&existingBill, bill.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bill not found with ID:", bill.Id)
		}
		return err
	}

	// 2. ตรวจสอบว่า invoice ถูกเปลี่ยนหรือไม่
	if existingBill.Invoice != bill.Invoice {
		// 3. ถ้าเปลี่ยน invoice, ตรวจสอบว่า invoice ใหม่ซ้ำหรือไม่
		var count int64
		if err := r.db.Model(&Bill_Header{}).
			Where("invoice = ? AND id != ?", bill.Invoice, bill.Id).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check invoice duplication: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("invoice already exists", bill.Invoice)
		}
	}

	// 4. อัปเดตข้อมูลทั้งหมดของ bill
	if err := r.db.Model(&Bill_Header{}).
		Where("id = ?", bill.Id).
		Save(bill).Error; err != nil {
		return fmt.Errorf("failed to update bill:", err)
	}

	return nil
}
func (r *billRepositoryDB) UpdateBillStatus(bill *Bill_Header) error {
	var existingBill Bill_Header

	// 1. โหลดบิลเดิม
	if err := r.db.First(&existingBill, bill.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bill not found with ID:", bill.Id)
		}
		return err
	}

	// 2. อัปเดตเฉพาะฟิลด์ที่ต้องการ (Status)
	existingBill.Status = bill.Status
	existingBill.Note = bill.Note

	// 3. บันทึกการเปลี่ยนแปลง
	if err := r.db.Save(&existingBill).Error; err != nil {
		return fmt.Errorf("failed to update bill:", err)
	}

	return nil
}
func (r *billRepositoryDB) UpdateInstallmentBillStatus(bill *model.Bill_Header_Installment) error {
	var existingBill model.Bill_Header_Installment

	// 1. โหลดบิลเดิม
	if err := r.db.First(&existingBill, bill.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bill not found with ID: ", bill.Id)
		}
		return err
	}

	// 2. อัปเดตเฉพาะฟิลด์ที่ต้องการ (Status)
	existingBill.Status = bill.Status
	existingBill.Note = bill.Note

	// 3. บันทึกการเปลี่ยนแปลง
	if err := r.db.Save(&existingBill).Error; err != nil {
		return fmt.Errorf("failed to update bill: ", err)
	}

	return nil
}
func (r *billRepositoryDB) UpdateBillFee(bill *Bill_Header) error {
	return r.db.Model(&model.Bill_Header{}).
		Where("id = ?", bill.Id).
		Updates(map[string]interface{}{
			"fee_amount":       bill.Fee_Amount,
			"remaining_amount": bill.Remaining_Amount,
			"late_day":         bill.Late_Day,
		}).Error
}

func (r *billRepositoryDB) GetAllBillDetails(billID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.Where("bill_header_id = ?", billID).Order("payment_date ASC").Find(&details).Error
	return details, err
}

func (r *billRepositoryDB) GetAllUnpaidBills() ([]model.Bill_Header, error) {
	var bills []model.Bill_Header

	// ใช้ Join กับ Bill_Details เพื่อตรวจสอบงวดที่ยัง unpaid
	err := r.db.
		Joins("JOIN bill_details ON bill_details.bill_header_id = bill_headers.id").
		Where("bill_details.status = ? AND bill_headers.status = ?  ", 0, 1). // 0 = ยังไม่จ่าย
		Order("id ASC").
		Group("bill_headers.id").
		Find(&bills).Error

	if err != nil {
		return nil, err
	}

	return bills, nil
}

func (r *billRepositoryDB) GetAllInstallments(billID uint) ([]Bill_Details, error) {
	var installments []Bill_Details
	if err := r.db.Where("bill_header_id = ?", billID).Where("status = 1").Find(&installments).Error; err != nil {
		return nil, err
	}
	return installments, nil
}

func (r *billRepositoryDB) GetBillWithUnpaidInstallments(billID uint) (*Bill_Header, error) {
	var bill Bill_Header
	err := r.db.Preload("BillDetails", "status = ?", 0).First(&bill, billID).Error
	return &bill, err
}

func (r *billRepositoryDB) GetUnpaidBillsBatch(limit, offset int) ([]model.Bill_Header, error) {
	var bills []model.Bill_Header
	err := r.db.Where("status = ?", 0).
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&bills).Error
	return bills, err
}

// ////////////////////////////////////////////
func (r *billRepositoryDB) GetUnpaidInstallmentsByBillIDs(billIDs []uint) ([]model.Bill_Details, error) {
	var details []model.Bill_Details
	err := r.db.
		Where("bill_header_installment_id IN ? AND status = ?", billIDs, 0).
		Find(&details).Error
	return details, err
}

func (r *billRepositoryDB) UpdateBillDetailsBatch(details []model.Bill_Details) error {
	if len(details) == 0 {
		return nil
	}
	// ใช้ transaction เพื่อ update หลาย row
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, inst := range details {
			if err := tx.Model(&model.Bill_Details{}).
				Where("id = ?", inst.Id).
				Updates(map[string]interface{}{
					"fee_amount":        inst.Fee_Amount,
					"installment_price": inst.Installment_Price,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *billRepositoryDB) UpdateBillsBatch(bills []*model.Bill_Header) error {
	if len(bills) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, b := range bills {
			if err := tx.Model(&model.Bill_Header{}).
				Where("id = ?", b.Id).
				Updates(map[string]interface{}{
					"fee_amount":       b.Fee_Amount,
					"remaining_amount": b.Remaining_Amount,
					"late_day":         b.Late_Day,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *billRepositoryDB) UpdateBillDetailBatch(installments []model.Bill_Details) error {
	if len(installments) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range installments {
			if err := tx.Save(&installments[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *billRepositoryDB) UpdateBillBatch(bills []*model.Bill_Header) error {
	if len(bills) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range bills {
			if err := tx.Model(bills[i]).Updates(bills[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *billRepositoryDB) GetAllBill(filter BillFilter, limit, offset int, bestProductIds []uint, sortOrder int) ([]Bill_Header, error) {
	var bills []Bill_Header
	query := r.db.Preload("BillDetails").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User")
	query = query.Where("bill_headers.status = ?", 2)

	// if filter.Status != nil {
	// 	query = query.Where("bill_headers.status = ?", *filter.Status)
	// }
	if len(bestProductIds) > 0 {
		order := "DESC"
		if sortOrder == 2 { // 2 = น้อยไปมาก
			order = "ASC"
		}
		query = query.
			Where("bill_headers.product_id IN ?", bestProductIds).
			Joins("JOIN (SELECT product_id, COUNT(*) AS total_sold FROM bill_headers WHERE status = 2 GROUP BY product_id) AS sales ON sales.product_id = bill_headers.product_id").
			Order("sales.total_sold " + order)
	}

	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}

	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&bills).Error
	return bills, err
}

func (r *billRepositoryDB) CountBills(filter BillFilter) (int64, error) {
	var count int64
	query := r.db.Model(&model.Bill_Header{})
	query = query.Where("bill_headers.status = ?", 2)

	if len(filter.Invs) > 0 {
		for _, inv := range filter.Invs {
			if inv != "" {
				query = query.Where("invoice ILIKE ?", "%"+inv+"%")
			}
		}
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

func (r *billRepositoryDB) GetBestSellingProducts(limit int) ([]BestSellingProduct, error) {
	var results []BestSellingProduct

	err := r.db.Model(&Bill_Header{}).
		Select("product_id, COUNT(*) as total_sold").
		Where("status = ?", 1). // สมมติ status=1 คือจ่ายแล้ว
		Group("product_id").
		Order("total_sold DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *billRepositoryDB) SumPaidAndUnpaidCounts(filter BillFilter) (paidCount int64, unpaidCount int64, err error) {
	applyFilters := func(q *gorm.DB) *gorm.DB {
		if len(filter.Invs) > 0 {
			for _, inv := range filter.Invs {
				if inv != "" {
					q = q.Where("invoice ILIKE ?", "%"+inv+"%")
				}
			}
		}
		if filter.DateFrom != nil {
			q = q.Where("created_at >= ?", *filter.DateFrom)
		}
		if filter.DateTo != nil {
			endDate := filter.DateTo.AddDate(0, 0, 1)
			q = q.Where("created_at < ?", endDate)
		}
		return q
	}

	// ✅ Paid (status=2)
	if err = applyFilters(r.db.Model(&Bill_Header{})).
		Where("status = ?", 2).
		Count(&paidCount).Error; err != nil {
		return
	}

	// ✅ Unpaid (status=1)
	if err = applyFilters(r.db.Model(&Bill_Header{})).
		Where("status = ?", 1).
		Count(&unpaidCount).Error; err != nil {
		return
	}

	return
}

func (r *billRepositoryDB) GetBillSummary(filter BillFilter) (summary BillSummary, err error) {
	query := r.db.Model(&Bill_Header{})

	if len(filter.Invs) > 0 {
		for _, inv := range filter.Invs {
			if inv != "" {
				query = query.Where("invoice ILIKE ?", "%"+inv+"%")
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	err = query.Select(`
		COALESCE(SUM(CASE WHEN status = 2 THEN paid_amount ELSE 0 END),0) AS paid_total,
		COALESCE(SUM(CASE WHEN status = 1 THEN remaining_amount ELSE 0 END),0) AS unpaid_total,
		COUNT(CASE WHEN status = 2 THEN 1 END) AS paid_count,
		COUNT(CASE WHEN status = 1 THEN 1 END) AS unpaid_count
	`).Scan(&summary).Error

	return
}

func (r *billRepositoryDB) GetLastHpcByYear(yearSuffix string) (string, error) {
	var lastSKU string
	err := r.db.Model(&model.Bill_Header_Installment{}).
		Select("invoice").
		// หาเฉพาะปี (ไม่สนใจเดือน)
		Where("invoice ILIKE ?", fmt.Sprintf("HPC/%%-%s-%%", yearSuffix)).
		Order("invoice DESC").
		Limit(1).
		Scan(&lastSKU).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	return lastSKU, nil
}

func GenerateHpc(r BillRepository) (string, error) {
	now := time.Now()
	month := fmt.Sprintf("%02d", int(now.Month()))
	year := now.Year() + 543
	yearSuffix := fmt.Sprintf("%02d", year%100)

	lastSKU, err := r.GetLastHpcByYear(yearSuffix)
	if err != nil {
		return "", err
	}

	nextSeq := 1
	if lastSKU != "" {
		// ตัวอย่าง lastSKU = "BI/09-68-0001"
		parts := strings.Split(lastSKU, "-")
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid SKU format: %s", lastSKU)
		}
		num, err := strconv.Atoi(parts[2]) // "0001"
		if err != nil {
			return "", fmt.Errorf("invalid sequence number in SKU: %s", lastSKU)
		}
		nextSeq = num + 1
	}

	// คืนค่าในรูปแบบ BI/09-68-0001
	return fmt.Sprintf("HPC/%s-%s-%04d", month, yearSuffix, nextSeq), nil
}

func (r *billRepositoryDB) CreateInstallmentBill(bill *model.Bill_Header_Installment) (*model.Bill_Header_Installment, error) {
	// เช็คเฉพาะ invoice ก็พอ
	var existing model.Bill_Header_Installment
	if err := r.db.Where("invoice = ?", bill.Invoice).First(&existing).Error; err == nil {
		return nil, gorm.ErrDuplicatedKey
	}

	if err := r.db.Create(bill).Error; err != nil {
		return nil, err
	}
	return bill, nil
}

func (r *billRepositoryDB) CreateInstallmentBillDetails(details []model.Bill_Details_Installment) error {
	if len(details) == 0 {
		return errors.New("no bill details provided")
	}

	// ไม่ต้องเช็ค duplicated จาก Id เพราะ DB auto generate เอง
	if err := r.db.Create(&details).Error; err != nil {
		return err
	}
	return nil
}

func (r *billRepositoryDB) GetUnpaidBillInstallments(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_installment_id = ? AND status = 0", billID).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidBillInstallments3(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		Where("bill_header_installment_id = ? AND status IN ?", billID, []int{0, 2}).
		Order("payment_date ASC, id ASC").
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidBillInstallments2(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_installment_id = ? ", billID).
		Order("id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) UpdateInstallmentBillDetail(installments []model.Bill_Details_Installment) error {
	// แก้ไข: loop แบบ index เพื่อให้เราได้ pointer ไปที่ element ใน slice จริงๆ
	for i := range installments {
		if err := r.db.Save(&installments[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *billRepositoryDB) UpdateBillInstallment(bill *model.Bill_Header_Installment) error {
	// ✅ อัปเดตเฉพาะ field ที่ต้องการ
	updateData := map[string]interface{}{
		"interest_amount":        bill.Interest_Amount,
		"total_interest_amount":        bill.Total_Interest_Amount,

		"net_installment":        bill.Net_installment,
		"remaining_amount":       bill.Remaining_Amount,
		"paid_amount":            bill.Paid_Amount,
		"credit_balance":         bill.Credit_Balance,
		"total_installments":     bill.Total_Installments,
		"paid_installments":      bill.Paid_Installments,
		"status":                 bill.Status,
		"remaining_installments": bill.Remaining_Installments,
		"next_due_date":          bill.NextDueDate,
		"last_renew_date":        bill.LastRenewDate,
		"late_day": bill.Late_Day,
		"fee_amount":bill.Fee_Amount,
		// "total_price":     <-- ไม่ใส่!
	}
	fmt.Println("bill.Id:", bill.Id)
	return r.db.Model(&model.Bill_Header_Installment{}).Where("id = ?", bill.Id).Updates(updateData).Error
}

func (r *billRepositoryDB) GetInstallmentBillById(id uint) (*model.Bill_Header_Installment, error) {
	var bill model.Bill_Header_Installment
	err := r.db.
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("BillDetailsInstallment"). // ✅ preload งวดด้วย
		Preload("User").
		First(&bill, id).Error
	return &bill, err
}
func (r *billRepositoryDB) GetInstallmentDetailsByBillID(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		Where("bill_header_installment_id = ?", billID).
		Preload("Bill_Header_Installment").
		Preload("Bill_Header_Installment.Member").
		Preload("Bill_Header_Installment.User").
		Preload("Bill_Header_Installment.Product").
		Preload("Bill_Header_Installment.Product.Category").
		Find(&details).Error
	return details, err
}

func (r *billRepositoryDB) GetPaidInstallmentsBill(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		Where("bill_header_installment_id = ? AND status = 1", billID).
		Order("payment_date ASC").
		Find(&details).Error
	return details, err
}

func (r *billRepositoryDB) GetAllInstallmentUnpaidBills() ([]model.Bill_Header_Installment, error) {
	var bills []model.Bill_Header_Installment

	// ใช้ Joins กับ Bill_Details_Installment เพื่อตรวจสอบงวดที่ยัง unpaid
	err := r.db.
		Joins("JOIN bill_details_installments ON bill_details_installments.bill_header_installment_id = bill_header_installments.id"). // ใช้ชื่อ table ที่ถูกต้อง
		Where("bill_details_installments.status = ?", 0).                                                                              // 0 = ยังไม่จ่าย
		Group("bill_header_installments.id").                                                                                          // Group by Bill_Header_Installment ID
		Preload("BillDetailsInstallment", "status = ?", 0).                                                                            // Preload งวดที่ยัง unpaid
		Find(&bills).Error

	if err != nil {
		return nil, err
	}

	return bills, nil
}

func (r *billRepositoryDB) UpdateInstallmentBill(bill *model.Bill_Header_Installment) error {
	return r.db.Save(bill).Error
}

func (r *billRepositoryDB) UpdateBillFeeInstallment(bill *model.Bill_Header_Installment) error {
	return r.db.Model(&model.Bill_Header_Installment{}).
		Where("id = ?", bill.Id).
		Updates(map[string]interface{}{
			"fee_amount":       bill.Fee_Amount,
			"remaining_amount": bill.Remaining_Amount,
			"late_day":         bill.Late_Day,
		}).Error
}
func (r *billRepositoryDB) UpdateBillInstallmentDetail(installments []model.Bill_Details_Installment) error {
	// แก้ไข: loop แบบ index เพื่อให้เราได้ pointer ไปที่ element ใน slice จริงๆ
	for i := range installments {
		if err := r.db.Save(&installments[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
func (r *billRepositoryDB) GetInstallmentBillDetailById(id uint) (*model.Bill_Details_Installment, error) {
	var details model.Bill_Details_Installment
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("id = ? AND status = 0", id).
		Order("payment_date ASC ").
		First(&details).Error
	return &details, err
}

func (r *billRepositoryDB) UpdateInstallmentSingleBillDetail(detail *model.Bill_Details_Installment) error {
	return r.db.Save(detail).Error
}

func (r *billRepositoryDB) CountInstallmentBills(filter BillFilter) (int64, error) {
	var count int64
	query := r.db.Model(&model.Bill_Header_Installment{})

	// ✅ กรองสถานะ
	query = query.Where("status = ?", 2)

	// ✅ กรอง Invoice
	if len(filter.Invs) > 0 {
		var cleanInvs []string
		for _, inv := range filter.Invs {
			if inv != "" {
				cleanInvs = append(cleanInvs, inv)
			}
		}
		if len(cleanInvs) > 0 {
			for i, inv := range cleanInvs {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}

	// ✅ กรองวันที่
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

func (r *billRepositoryDB) GetInstallmentAllBill(
	filter BillFilter,
	limit, offset int,
	bestProductIds []uint,
	sortOrder int,
) ([]model.Bill_Header_Installment, error) {
	var bills []model.Bill_Header_Installment

	query := r.db.Model(&model.Bill_Header_Installment{}).
		Preload("BillDetailsInstallment").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User")

	// ✅ กรองเฉพาะ status = 2
	query = query.Where("bill_header_installments.status = ?", 2)

	// ✅ กรณีมี bestProductIds (สินค้าขายดี)
	if len(bestProductIds) > 0 {
		order := "DESC"
		if sortOrder == 2 {
			order = "ASC"
		}

		salesSubQuery := r.db.
			Table("bill_header_installments").
			Select("product_id, COUNT(*) AS total_sold").
			Where("status = ?", 2).
			Group("product_id")

		query = query.
			Where("bill_header_installments.product_id IN ?", bestProductIds).
			Joins("JOIN (?) AS sales ON sales.product_id = bill_header_installments.product_id", salesSubQuery).
			Order("sales.total_sold " + order)
	}

	// ✅ กรอง Invoice
	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}
	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	// ✅ กรองวันที่
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	// ✅ Limit / Offset
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&bills).Error
	return bills, err
}

func (r *billRepositoryDB) GetBestSellingInstallmentsProducts(limit int) ([]BestSellingProduct, error) {
	var results []BestSellingProduct

	err := r.db.Model(&model.Bill_Header_Installment{}).
		Select("product_id, COUNT(*) as total_sold").
		Where("status = ?", 2).
		Group("product_id").
		Order("total_sold DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *billRepositoryDB) SumInstallmentPaidAndUnpaidCounts(filter BillFilter) (paidCount int64, unpaidCount int64, err error) {
	applyFilters := func(q *gorm.DB) *gorm.DB {
		if len(filter.Invs) > 0 {
			for _, inv := range filter.Invs {
				if inv != "" {
					q = q.Where("invoice ILIKE ?", "%"+inv+"%")
				}
			}
		}
		if filter.DateFrom != nil {
			q = q.Where("created_at >= ?", *filter.DateFrom)
		}
		if filter.DateTo != nil {
			endDate := filter.DateTo.AddDate(0, 0, 1)
			q = q.Where("created_at < ?", endDate)
		}
		return q
	}

	// ✅ Paid (status=2)
	if err = applyFilters(r.db.Model(&model.Bill_Header_Installment{})).
		Where("status = ?", 2).
		Count(&paidCount).Error; err != nil {
		return
	}

	// ✅ Unpaid (status=1)
	if err = applyFilters(r.db.Model(&model.Bill_Header_Installment{})).
		Where("status = ?", 1).
		Count(&unpaidCount).Error; err != nil {
		return
	}

	return
}

func (r *billRepositoryDB) GetInstallmentBillSummary(filter BillFilter) (summary BillSummary, err error) {
	query := r.db.Model(&model.Bill_Header_Installment{})

	if len(filter.Invs) > 0 {
		for _, inv := range filter.Invs {
			if inv != "" {
				query = query.Where("invoice ILIKE ?", "%"+inv+"%")
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	err = query.Select(`
		COALESCE(SUM(CASE WHEN status = 2 THEN paid_amount ELSE 0 END),0) AS paid_total,
		COALESCE(SUM(CASE WHEN status = 1 THEN remaining_amount ELSE 0 END),0) AS unpaid_total,
		COUNT(CASE WHEN status = 2 THEN 1 END) AS paid_count,
		COUNT(CASE WHEN status = 1 THEN 1 END) AS unpaid_count
	`).Scan(&summary).Error

	return
}

func (r *billRepositoryDB) UpdateBill_Installments(bill *model.Bill_Header_Installment) error {
	var existingBill model.Bill_Header_Installment

	// 1. ตรวจสอบว่าบิลที่ต้องการแก้ไขมีอยู่หรือไม่
	if err := r.db.First(&existingBill, bill.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bill not found with ID: ", bill.Id)
		}
		return err
	}

	// 2. ตรวจสอบว่า invoice ถูกเปลี่ยนหรือไม่
	if existingBill.Invoice != bill.Invoice {
		// 3. ถ้าเปลี่ยน invoice, ตรวจสอบว่า invoice ใหม่ซ้ำหรือไม่
		var count int64
		if err := r.db.Model(&model.Bill_Header_Installment{}).
			Where("invoice = ? AND id != ?", bill.Invoice, bill.Id).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check invoice duplication: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("invoice  already exists", bill.Invoice)
		}
	}

	// 4. อัปเดตข้อมูลทั้งหมดของ bill
	if err := r.db.Model(&model.Bill_Header_Installment{}).
		Where("id = ?", bill.Id).
		Updates(bill).Error; err != nil {
		return fmt.Errorf("failed to update bill: ", err)
	}

	return nil
}

func (r *billRepositoryDB) GetUnPayAllBill(filter BillFilter, limit, offset int, sortOrder int) ([]Bill_Header, error) {
	var bills []Bill_Header
	query := r.db.Preload("BillDetails").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User")

	query = query.Where("bill_headers.status in ?", []int{0, 1})
	order := "DESC"
	if sortOrder == 2 {
		order = "ASC"
	}
	query = query.Order("id " + order)

	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&bills).Error
	return bills, err
}
func (r *billRepositoryDB) CountUnpayBills(filter BillFilter) (int64, error) {
	var count int64
	query := r.db.Model(&model.Bill_Header{})

	// ✅ ใส่เงื่อนไขให้เหมือน GetUnPayAllBill
	query = query.Where("status IN ?", []int{0, 1})
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }

	// ✅ กรอง invoice
	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}
	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	// ✅ กรองวันที่
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

func (r *billRepositoryDB) GetInstallmentAllBillUnpay(
	filter BillFilter,
	limit, offset int,
	sortOrder int,
) ([]model.Bill_Header_Installment, error) {
	var bills []model.Bill_Header_Installment

	query := r.db.Model(&model.Bill_Header_Installment{}).
		Preload("BillDetailsInstallment").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User")
	query = query.Where("bill_header_installments.status in ?", []int{0, 1})

	// query = query.Where("bill_header_installments.status = ?", 1)
	order := "DESC"
	if sortOrder == 2 {
		order = "ASC"
	}
	query = query.Order("id " + order)

	// ✅ กรอง Invoice
	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	// ✅ กรองวันที่
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	// ✅ Limit / Offset
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	// ✅ ดึงข้อมูล
	err := query.Find(&bills).Error
	return bills, err
}

func (r *billRepositoryDB) CountInstallmentBillsUnpay(filter BillFilter) (int64, error) {
	var count int64
	query := r.db.Model(&model.Bill_Header_Installment{})

	// ✅ ต้องใส่ status เหมือน GetInstallmentAllBillUnpay
	query = query.Where("status IN ?", []int{0, 1})
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	// ✅ กรอง Invoice ให้เหมือนกัน
	var cleanInvs []string
	for _, inv := range filter.Invs {
		if inv != "" {
			cleanInvs = append(cleanInvs, inv)
		}
	}
	if len(cleanInvs) > 0 {
		for i, inv := range cleanInvs {
			pattern := "%" + inv + "%"
			if i == 0 {
				query = query.Where("invoice ILIKE ?", pattern)
			} else {
				query = query.Or("invoice ILIKE ?", pattern)
			}
		}
	}

	// ✅ กรองวันที่
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	// ✅ นับ
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *billRepositoryDB) GetUnpaidInstallmentsByDate() ([]Bill_Details1, error) {
	var details []Bill_Details1

	today := time.Now().Format("2006-01-02")
	err := r.db.
		Table("bill_details").
		Preload("BillHeader").
		Preload("BillHeader.Member").  // ✅ preload Member ด้วย
		Preload("BillHeader.Product"). // (ถ้าใช้ Product)
		Preload("BillHeader.User").    // (ถ้าใช้ User)
		Where("status = 0 AND DATE(payment_date) <= ?", today).
		Order("payment_date ASC, id ASC").
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidInstallBillmentsByDate() ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment

	// ดึงข้อมูลทั้งหมดที่ status=0 และ payment_date <= วันนี้ (รวม today และ overdue)
	today := time.Now().Format("2006-01-02")

	// err := r.db.
	err := r.db.
		Preload("Bill_Header_Installment").
		Preload("Bill_Header_Installment.Member").
		Preload("Bill_Header_Installment.User").
		Preload("Bill_Header_Installment.Product.Category").
		Where("status = 0 AND DATE(payment_date) <= ?", today).
		Order("payment_date ASC, id ASC").
		Find(&details).Error

	// Where("status = 0 AND DATE(payment_date) <= ?", today).
	// Order("payment_date ASC, id ASC").
	// Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidBill(userId string) ([]Bill_Details1, error) {
	var details []Bill_Details1

	// err := r.db.
	// 	Debug().
	err := r.db.
		Debug().
		Table("bill_details as bd").
		Select("bd.*").
		Joins("JOIN bill_headers as bh ON bd.bill_header_id = bh.id").
		Joins("JOIN members as m ON bh.member_id = m.id").
		Where("m.user_id = ?", userId).
		Where("bd.status = 0").
		Where("bh.status = 1").
		// Where("bd.payment_date <= NOW() - INTERVAL '1 minute'").
		Order("bd.payment_date ASC, bd.id ASC").
		Preload("BillHeader").
		Preload("BillHeader.Member").
		Preload("BillHeader.User").
		Preload("BillHeader.Product").
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidInstallmentBill(userId string) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment

	err := r.db.
		Debug().
		Table("bill_details_installments as bd").
		Select("bd.*").
		Joins("JOIN bill_header_installments as bh ON bd.bill_header_installment_id = bh.id").
		Joins("JOIN members as m ON bh.member_id = m.id").
		Where("m.user_id = ?", userId).
		// Where("bd.bill_header_installment_id = ?", billID).
		Where("bd.status = 0").
		Where("bh.status = 1").
		// Where("bd.payment_date <= NOW() - INTERVAL '1 minute'").
		Order("bd.payment_date ASC, bd.id ASC").
		Preload("Bill_Header_Installment").
		Preload("Bill_Header_Installment.Member").
		Preload("Bill_Header_Installment.User").
		Preload("Bill_Header_Installment.Product").
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetUnpaidInstallmentsInvoice(invoice string) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("invoice = ? AND status = 0", invoice).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}
func (r *billRepositoryDB) GetPaidInstallmentsInvoice(invoice string) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		Where("invoice = ? AND status = 1", invoice).
		Order("payment_date ASC").
		Find(&details).Error
	return details, err
}
func (r *billRepositoryDB) GetBillByInvoice(id uint, detail uint) (*Bill_Header, error) {
	var bill Bill_Header
	err := r.db.Preload("BillDetails").
		Preload("Member").
		Preload("Product").
		Preload("Product.Category").
		Preload("User").
		First(&bill, id).Error
	return &bill, err
}
func (r *billRepositoryDB) GetUnpaidInstallments1(billID uint, detailID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_id = ? AND status = 0 AND id = ?", billID, detailID).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetPaidInstallments1(billID uint) ([]Bill_Details, error) {
	var details []Bill_Details
	err := r.db.
		Debug().
		Where("bill_header_id = ? AND status = 1", billID).
		Order("payment_date ASC").
		Find(&details).Error
	return details, err
}
func (r *billRepositoryDB) GetInstallmentCounts(billID uint) (int64, int64, error) {
	var total int64
	var paid int64

	// นับงวดทั้งหมด
	if err := r.db.Model(&Bill_Details{}).
		Where("bill_header_id = ?", billID).
		Count(&total).Error; err != nil {
		return 0, 0, err
	}

	// นับงวดที่จ่ายแล้ว (status = 1)
	if err := r.db.Model(&Bill_Details{}).
		Where("bill_header_id = ? AND status = 1", billID).
		Count(&paid).Error; err != nil {
		return 0, 0, err
	}

	return total, paid, nil
}

func (r *billRepositoryDB) GetUnpaidBillInstallments1(billID uint, detailID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		// Debug(). // 👈 เปิดดู SQL ตรงนี้
		Where("bill_header_installment_id = ? AND status = 0 AND id = ?", billID, detailID).
		Order("payment_date ASC,id ASC"). // ใช้คอมม่าในการเรียงลำดับหลายคอลัมน์
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetPaidInstallmentsBill1(billID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment
	err := r.db.
		Where("bill_header_installment_id = ? AND status = 1  ", billID).
		Order("payment_date ASC").
		Find(&details).Error
	return details, err
}
func (r *billRepositoryDB) GetpaidBill(billID uint, detailID uint) ([]Bill_Details1, error) {
	var details []Bill_Details1

	// err := r.db.
	// 	Debug().
	err := r.db.
		Debug().
		Table("bill_details as bd").
		Select("bd.*").
		Joins("JOIN bill_headers as bh ON bd.bill_header_id = bh.id").
		Joins("JOIN members as m ON bh.member_id = m.id").
		Where("bh.id = ? AND bd.id = ?", billID, detailID).
		// Where("bd.status = 1").
		// Where("bh.status = 2").
		Order("bd.payment_date ASC, bd.id ASC").
		Preload("BillHeader").
		Preload("BillHeader.Member").
		Preload("BillHeader.User").
		Preload("BillHeader.Product").
		Find(&details).Error

	return details, err
}

func (r *billRepositoryDB) GetpaidInstallmentBill(billID uint, detailID uint) ([]model.Bill_Details_Installment, error) {
	var details []model.Bill_Details_Installment

	err := r.db.
		Debug().
		Table("bill_details_installments as bd").
		Select("bd.*").
		Joins("JOIN bill_header_installments as bh ON bd.bill_header_installment_id = bh.id").
		// 👇 ลองเปลี่ยนเป็น OR เพื่อเช็กว่าเงื่อนไขไหน fail
		Where("bh.id = ? AND bd.id = ?", billID, detailID).
		// 🧪 ชั่วคราว: เอา status ออก เพื่อดูว่ามีข้อมูลหรือไม่
		// Where("bd.status = 1").
		// Where("bh.status = 2").
		Order("bd.payment_date ASC, bd.id ASC").
		Preload("Bill_Header_Installment").
		Preload("Bill_Header_Installment.Member").
		Preload("Bill_Header_Installment.User").
		Preload("Bill_Header_Installment.Product").
		Find(&details).Error

	return details, err
}
func (r *billRepositoryDB) SumPaidAmountByStatus2(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header{}).
		Select("COALESCE(SUM(paid_amount - fee_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 2)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
func (r *billRepositoryDB) SumFeeByStatus2(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header{}).
		Select("COALESCE(SUM(fee_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 2)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ใน billRepositoryDB.go
func (r *billRepositoryDB) SumPaidAmountByStatus1(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header{}).
		Select("COALESCE(SUM(paid_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 1)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *billRepositoryDB) SumPaidAmountByInstallmentStatus1(filter BillFilter) (float64, error) {
	// func (r *billRepositoryDB) SumPaidAmountByStatus1(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header_Installment{}).
		Select("COALESCE(SUM(paid_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 1)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
func (r *billRepositoryDB) SumPaidAmountByInstallmentStatus2(filter BillFilter) (float64, error) {
	// func (r *billRepositoryDB) SumPaidAmountByStatus1(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header_Installment{}).
		Select("COALESCE(SUM(paid_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 2)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_header_installments.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
func (r *billRepositoryDB) SumFeeInstallmentByStatus2(filter BillFilter) (float64, error) {
	var total float64
	query := r.db.Model(&model.Bill_Header_Installment{}).
		Select("COALESCE(SUM(fee_amount), 0)")

	// ✅ ใส่เงื่อนไขให้เหมือน filter
	query = query.Where("status = ?", 2)
	if len(filter.NameOrPhones) > 0 {
		query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id")
		for i, val := range filter.NameOrPhones {
			if val == "" {
				continue
			}
			pattern := "%" + val + "%"
			if i == 0 {
				query = query.Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			} else {
				query = query.Or("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
			}
		}
	}

	// if filter.NameOrPhone != "" {
	// 	pattern := "%" + filter.NameOrPhone + "%"
	// 	query = query.Joins("JOIN members as m ON bill_headers.member_id = m.id").
	// 		Where("m.full_name ILIKE ? OR m.tel ILIKE ?", pattern, pattern)
	// }
	if len(filter.Invs) > 0 {
		for i, inv := range filter.Invs {
			if inv != "" {
				pattern := "%" + inv + "%"
				if i == 0 {
					query = query.Where("invoice ILIKE ?", pattern)
				} else {
					query = query.Or("invoice ILIKE ?", pattern)
				}
			}
		}
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		endDate := filter.DateTo.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *billRepositoryDB) GetAllUnpaid10DayBills() ([]model.Bill_Header_Installment, error) {
	var bills []model.Bill_Header_Installment
	err := r.db.Preload("BillDetailsInstallment").
		Where("installment_day = ? AND status = ? AND term_type = ?", 10, 1, 1).
		Find(&bills).Error
	if err != nil {
		return nil, err
	}
	return bills, nil
}

// ใหม่:
func (r *billRepositoryDB) UpdateInstallmentBillDetail1(installment *model.Bill_Details_Installment) error {
	return r.db.Save(installment).Error
}


func (r *billRepositoryDB) CreateInstallmentDetail1(detail *model.Bill_Details_Installment) error {
	if detail == nil {
		return errors.New("invalid bill detail")
	}
	return r.db.Create(detail).Error
}


func (r *billRepositoryDB) UpdateInstallmentDetail(detail *model.Bill_Details_Installment) error {
    // สร้าง map เพื่อระบุฟิลด์ที่จะอัปเดตอย่างชัดเจน
    updates := map[string]interface{}{
        "status":            detail.Status,
        "fee_amount":        detail.Fee_Amount,
        "installment_price": detail.Installment_Price,
        "paid_amount":       detail.Paid_Amount,
    }
    // ใช้ Updates เพื่ออัปเดตหลายฟิลด์พร้อมกัน โดยระบุ ID ให้ชัดเจน
    return r.db.Model(&model.Bill_Details_Installment{}).Where("id = ?", detail.Id).Updates(updates).Error
}

func (r *billRepositoryDB) HasInterestInPeriod(billID uint, lastRenew, nextDue time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&model.Bill_Details_Installment{}).
		Where("bill_header_installment_id = ? AND is_interest_only = ? AND payment_date BETWEEN ? AND ?",
			billID, true, lastRenew, nextDue).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ใน billRepositoryDB.go
func (r *billRepositoryDB) CloseOldInterestInstallments(billID uint) error {
	return r.db.Model(&model.Bill_Details_Installment{}).
		Where("bill_header_installment_id = ? AND is_interest_only = ? AND status != ?", billID, true, 2).
		Update("status", 2).Error
}
