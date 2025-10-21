package service

import (
	"errors"
	"fmt"
	"log"
	"math"
	"rrmobile/model"
	"rrmobile/respository"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type billService struct {
	billRepository        respository.BillRepository
	productRepository     respository.ProductRepository
	fineRepositoty        respository.FineRepository
	installmentRepository respository.InstallmentRepository
}

func NewBillService(billRepository respository.BillRepository, productRepository respository.ProductRepository, fineRepositoty respository.FineRepository, installmentRepository respository.InstallmentRepository) BillService {
	return &billService{billRepository: billRepository, productRepository: productRepository, fineRepositoty: fineRepositoty, installmentRepository: installmentRepository}
}

func (s *billService) CreateBill(request NewBillHeader) (*Bill_HeaderResponse, error) {
	product, err := s.productRepository.GetProductByID(request.ProductId)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if request.MemberId == 0 {
		return nil, errors.New("member id is required")
	}
	basePrice := product.Price

	finalPrice := basePrice + (basePrice * float64(request.Extra_Percent) / 100)

	downPayment := finalPrice * float64(request.Down_Percent) / 100
	remaining := finalPrice - downPayment
	roundedRemaining := math.Round(remaining)

	// installmentPrice := remaining / float64(request.Installments_Month)
	installmentPrice := math.Round(remaining / float64(request.Installments_Month))

	billNum, err := respository.GenerateInv(s.billRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SKU: ", err)
	}
	billHeader := &respository.Bill_Header{
		Invoice:            billNum,
		MemberId:           request.MemberId,
		User_Id:            request.User_Id,
		ProductId:          request.ProductId,
		Extra_Percent:      request.Extra_Percent,
		Down_Percent:       request.Down_Percent,
		Installments_Month: request.Installments_Month,
		Net_installment:    installmentPrice,
		Total_Price:        finalPrice,
		Remaining_Amount:   roundedRemaining,
		Total_Installments: request.Installments_Month,
		Status:             1,
	}

	createdBill, err := s.billRepository.CreateBill(billHeader)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	startDate := time.Now().In(loc)

	var details []respository.Bill_Details
	for i := 1; i <= request.Installments_Month; i++ {
		details = append(details, respository.Bill_Details{
			Bill_HeaderId:     createdBill.Id,
			Installment_Price: installmentPrice,
			Status:            0,
			Payment_Date:      startDate.AddDate(0, i, 0), // เพิ่มทีละเดือน
			Payment_No:        fmt.Sprintf("%d", i),
		})
	}

	if err := s.billRepository.CreateBillDetails(details); err != nil {
		return nil, err
	}

	resp := &Bill_HeaderResponse{
		Id:                 createdBill.Id,
		MemberId:           createdBill.MemberId,
		ProductId:          createdBill.ProductId,
		Total_Price:        createdBill.Total_Price,
		Net_installment:    createdBill.Net_installment,
		Total_Installments: createdBill.Total_Installments,
	}

	return resp, nil
}

func (s *billService) PayInstallment(billID uint, detailID uint, amount float64) ([]InstallmentPayResult, error) {
	// 1. ดึง Bill_Header
	bill, err := s.billRepository.GetBillById(billID)
	if err != nil {
		return nil, errors.New("bill not found")
	}

	maxPayable := float64(bill.Total_Installments)*bill.Net_installment + bill.Fee_Amount + bill.Credit_Balance

	remainingBill := maxPayable - float64(bill.Paid_Amount)
	if remainingBill < 0 {
		remainingBill = 0
	}

	// 2. ดึงเฉพาะ Bill_Details ที่ยังไม่จ่าย
	installments, err := s.billRepository.GetUnpaidInstallments1(billID, detailID)
	if err != nil {
		return nil, errors.New("cannot get installments")
	}
	if len(installments) == 0 {
		return nil, errors.New("all installments already paid")
	}

	remainingAmount := amount
	carryCredit := bill.Credit_Balance // เครดิตสะสมที่มีจากหัวบิล
	if remainingAmount > float64(bill.Remaining_Amount)+carryCredit {
		return nil, fmt.Errorf("ยอดชำระ %.2f เกินยอดคงเหลือของบิลและเครดิต %.2f", remainingAmount, float64(bill.Remaining_Amount)+carryCredit)
	}

	// if remainingAmount > float64(bill.Remaining_Amount) {
	// 	return nil, fmt.Errorf("ยอดชำระ %.2f เกินยอดคงเหลือของบิล %.2f", remainingAmount, float64(bill.Remaining_Amount))
	// }

	// **ประกาศ results slice ก่อนใช้**
	results := make([]InstallmentPayResult, 0, len(installments))

	// 3. ใช้เงิน + เครดิต ไปปิดงวดตามลำดับ
	for i := 0; i < len(installments) && (remainingAmount > 0 || carryCredit > 0); i++ {
		inst := &installments[i]
		unpaid := (inst.Installment_Price - inst.Paid_Amount)
		if unpaid <= 0 {
			continue
		}

		totalAvailable := remainingAmount + carryCredit
		if totalAvailable <= 0 {
			break
		}

		var result InstallmentPayResult
		result.InstallmentNo = i + 1

		// ---- Case A ----
		if totalAvailable == unpaid {
			inst.Paid_Amount = inst.Installment_Price
			inst.Status = 1
			inst.UpdatedAt = time.Now()
			if remainingAmount >= unpaid {
				remainingAmount -= unpaid
			} else {
				carryCredit -= unpaid - remainingAmount
				remainingAmount = 0
			}
			inst.Credit_Balance = 0

			result.Case = "A"
			result.Message = fmt.Sprintf("จ่ายพอดี งวด %d", i+1)
			result.CreditLeft = 0
			result.PaidAmount = inst.Paid_Amount

			// ---- Case B ----
		} else if totalAvailable > unpaid {
			inst.Paid_Amount = inst.Installment_Price
			inst.Status = 1
			inst.UpdatedAt = time.Now()

			over := totalAvailable - unpaid
			carryCredit = over
			remainingAmount = 0

			inst.Credit_Balance = carryCredit
			result.Case = "B"
			result.Message = fmt.Sprintf("จ่ายเกิน สร้างเครดิตใหม่ %.2f", carryCredit)
			result.CreditLeft = carryCredit
			result.PaidAmount = inst.Paid_Amount

			// ---- Case C ----
		} else if totalAvailable >= unpaid {
			inst.Paid_Amount = inst.Installment_Price
			inst.Status = 1
			inst.UpdatedAt = time.Now()

			need := unpaid - remainingAmount
			carryCredit -= need
			if carryCredit < 0 {
				carryCredit = 0
			}
			remainingAmount = 0

			inst.Credit_Balance = 0
			result.Case = "C"
			result.Message = fmt.Sprintf("ใช้เครดิต ปิดงวด เครดิตเหลือ %.2f", carryCredit)
			result.CreditLeft = carryCredit
			result.PaidAmount = inst.Paid_Amount

			// ---- Case D ----
		} else {
			if remainingAmount+carryCredit >= unpaid {
				inst.Paid_Amount = inst.Installment_Price
				inst.Status = 1
				inst.UpdatedAt = time.Now()

				over := remainingAmount + carryCredit - unpaid
				carryCredit = over
				remainingAmount = 0
				inst.Credit_Balance = 0

				result.Case = "D"
				result.Message = fmt.Sprintf("ปิดงวดด้วยเครดิตทั้งหมด เครดิตเหลือ %.2f", carryCredit)
				result.CreditLeft = carryCredit
				result.PaidAmount = inst.Paid_Amount
			}
		}

		results = append(results, result)
	}

	if err := s.billRepository.UpdateBillDetail(installments); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	bill.Credit_Balance = carryCredit
	paidInstallments, err := s.billRepository.GetPaidInstallments1(billID)
	if err != nil {
		return nil, errors.New("cannot get paid installments from DB")
	}
	fmt.Print(len(paidInstallments), "666")
	// bill.Paid_Installments = len(paidInstallments)

	// bill.Remaining_Installments = bill.Total_Installments - bill.Paid_Installments
	bill.Paid_Installments = len(paidInstallments)
	fmt.Print("bill.Paid_Installments", bill.Paid_Installments)
	fmt.Print("bill.Total_Installments", bill.Total_Installments)

	// bill.Remaining_Installments = bill.Total_Installments - bill.Paid_Installments
	kkk := bill.Total_Installments - bill.Paid_Installments
	fmt.Print("nigga how much?", kkk)
	// bill.Remaining_Installments -= kkk
	bill.Remaining_Installments = kkk
	if bill.Remaining_Installments < 0 {
		bill.Remaining_Installments = 0
	}
	fmt.Print("bill.Remaining_Installments", bill.Remaining_Installments)
	bill.Paid_Amount += int(amount - remainingAmount)
	if bill.Paid_Installments >= bill.Total_Installments {
		bill.Remaining_Amount = 0
		bill.Status = 2
	}

	// bill.Remaining_Amount -= amount - remainingAmount
	// if bill.Remaining_Amount <= 0 {
	// 	bill.Remaining_Amount = 0
	// 	bill.Status = 2
	// }

	if err := s.billRepository.UpdateBill(bill); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *billService) AutoApplyLateFees() error {
	start := time.Now()

	// โหลด timezone ไทย
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.FixedZone("Asia/Bangkok", 7*3600)
	}

	// 1) ดึงบิลทั้งหมดที่ยังไม่จ่าย
	bills, err := s.billRepository.GetAllUnpaidBills()
	if err != nil {
		return err
	}
	if len(bills) == 0 {
		return nil
	}

	// 2) ดึงค่าปรับรายวัน
	fine, err := s.fineRepositoty.GetFineById(1)
	if err != nil {
		return err
	}
	dailyFee := fine.FineAmount

	// today := time.Now().In(loc).Truncate(24 * time.Hour)
	updatedCount := int64(0)

	// 3) worker pool setup
	numJobs := len(bills)
	jobs := make(chan uint, numJobs)
	var wg sync.WaitGroup

	workers := 2
	for w := 0; w < workers; w++ {
		go func(workerID int) {
			for billID := range jobs {
				func(bid uint) {
					defer wg.Done()

					bill, err := s.billRepository.GetBillById(bid)
					if err != nil {
						return
					}

					installments, err := s.billRepository.GetUnpaidInstallments(bid)
					if err != nil {
						return
					}
					if len(installments) == 0 {
						return
					}

					changed := false
					graceDays := 15
					totalLateDays := 0 // ✅ เก็บรวม LateDays ของทุกงวดใน Bill นี้

					for i := range installments {
						inst := &installments[i]

						// ข้ามถ้างวดนี้จ่ายแล้ว
						if inst.Status != 0 {
							continue
						}

						// ล้างค่าปรับเก่าก่อนคำนวณใหม่
						inst.Installment_Price = round2(inst.Installment_Price - inst.Fee_Amount)
						bill.Fee_Amount = round2(bill.Fee_Amount - inst.Fee_Amount)
						bill.Remaining_Amount -= math.Round(inst.Fee_Amount)
						inst.Fee_Amount = 0

						todayDate := time.Now().In(loc)
						todayDate = time.Date(todayDate.Year(), todayDate.Month(), todayDate.Day(), 0, 0, 0, 0, loc)

						// วันที่ครบกำหนดจริง (วันจ่าย + วันผ่อนผัน)
						dueDate := inst.Payment_Date.In(loc).AddDate(0, 0, graceDays)
						dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)

						// ยังไม่ถึงกำหนด ไม่ต้องคิดค่าปรับ
						if !todayDate.After(dueDate) {
							continue
						}

						// คำนวณวันล่าช้า
						lateDays := int(todayDate.Sub(dueDate).Hours() / 24)
						if lateDays <= 0 {
							continue
						}

						// ✅ รวมวันล่าช้าสะสม
						totalLateDays += lateDays

						// คำนวณค่าปรับ
						fee := float64(lateDays) * dailyFee

						// อัปเดตข้อมูลใน installment
						inst.Fee_Amount = round2(fee)
						inst.Installment_Price = round2(inst.Installment_Price + fee)

						// log.Printf("Before: bill.Late_Day = %d, lateDays = %d", bill.Late_Day, lateDays)

						// ✅ ไม่อัปเดตตรงนี้ (จะอัปเดตหลัง loop)
						// bill.Late_Day = lateDays

						bill.Fee_Amount = round2(bill.Fee_Amount + fee)
						bill.Remaining_Amount += math.Round(fee)

						changed = true
						atomic.AddInt64(&updatedCount, 1)

						// log.Printf("BillID: %d, InstallmentID: %d, PaymentDate: %s, DueDate (after grace): %s, Today: %s, LateDays: %d",
						// 	bill.Id, inst.Id, inst.Payment_Date.In(loc), dueDate, today, lateDays)
					}

					// ✅ อัปเดต LateDay รวม (หลัง loop)
					if totalLateDays > 0 {
						bill.Late_Day = totalLateDays
						// log.Printf("✅ Total LateDays for BillID %d = %d", bill.Id, totalLateDays)
					}

					if !changed {
						return
					}

					// บันทึกลง DB
					if err := s.billRepository.UpdateBillDetail(installments); err != nil {
						// log.Printf("worker %d: failed to update installments for bill %d: %v", workerID, bill.Id, err)
					}
					if err := s.billRepository.UpdateBillFee(bill); err != nil {
						// log.Printf("worker %d: failed to update bill header for bill %d: %v", workerID, bill.Id, err)
					}

				}(billID)
			}
		}(w)
	}

	// ส่ง job เข้า worker
	for _, b := range bills {
		wg.Add(1)
		jobs <- b.Id
	}
	close(jobs)

	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("AutoApplyLateFees completed in %s, updated %d installments", elapsed, updatedCount)
	return nil
}

func round2(val float64) float64 {
	return math.Round(val*100) / 100
}
func (s *billService) AddExtraPayment(billID uint, installmentID uint, request UpdateAddExtraRequest) error {
	// 1. ดึง Bill_Header
	bill, err := s.billRepository.GetBillById(billID)
	if err != nil {
		return errors.New("bill not found")
	}

	// 2. ดึง Bill_Detail ตาม installmentID
	inst, err := s.billRepository.GetBillDetailById(installmentID)
	if err != nil {
		return errors.New("installment not found")
	}

	// 3. ตรวจสอบว่ายอดเงินที่จ่ายไม่เกินยอดที่เหลือของงวด
	remainingInst := inst.Installment_Price - inst.Paid_Amount
	if request.Paid_Amount > remainingInst {
		return fmt.Errorf("payment exceeds remaining amount of this installment: %.2f > %.2f", request.Paid_Amount, remainingInst)
	}

	// 4. ตรวจสอบว่ายอดเงินไม่เกินยอดคงเหลือของบิล
	if request.Paid_Amount > bill.Remaining_Amount {
		return fmt.Errorf("payment exceeds remaining amount of the bill: %.2f > %d", request.Paid_Amount, bill.Remaining_Amount)
	}

	// 5. เพิ่มยอด Paid_Amount ของงวด
	inst.Paid_Amount += request.Paid_Amount

	// 6. อัปเดตยอดรวมหัวบิล
	bill.Paid_Amount += int(request.Paid_Amount)
	bill.Remaining_Amount -= request.Paid_Amount

	// 7. ปรับ Status ของงวดถ้าปิด
	if inst.Paid_Amount >= inst.Installment_Price {
		inst.Status = 1                                                         // ปิดงวด
		inst.Credit_Balance = round2(inst.Paid_Amount - inst.Installment_Price) // ถ้ามีเงินเกิน เก็บเป็นเครดิต
	}

	// 8. อัปเดตจำนวนงวดที่ปิดในหัวบิล
	paidInstallments := 0
	for _, d := range bill.BillDetails {
		if d.Status == 1 || d.Id == inst.Id && inst.Status == 1 {
			paidInstallments++
		}
	}
	bill.Paid_Installments = paidInstallments
	bill.Remaining_Installments = bill.Total_Installments - paidInstallments

	// 9. ถ้า bill.Remaining_Amount <= 0 → ปิดบิล
	if bill.Remaining_Amount <= 0 {
		bill.Status = 2 // ปิดบิล
	}

	// 10. Save DB
	if err := s.billRepository.UpdateBill(bill); err != nil {
		return err
	}
	if err := s.billRepository.UpdateSingleBillDetail(inst); err != nil {
		return err
	}

	return nil
}

func (s *billService) GetAllBill(
	invs []string,
	dateFrom, dateTo *time.Time,
	page, limit int,
	sortOrder int, // <-- เพิ่มตรงนี้
) (*PaginationResponseBill, error) {

	if page < 1 {
		page = 1
	}

	filter := respository.BillFilter{Invs: invs, DateFrom: dateFrom, DateTo: dateTo}

	total, err := s.billRepository.CountBills(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count bills:", err)
	}
	paid, unpaid, err := s.billRepository.SumPaidAndUnpaidCounts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize bills: ", err)
	}
	sumPaid, err := s.billRepository.SumPaidAmountByStatus2(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}
	feeAmount, err := s.billRepository.SumFeeByStatus2(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}

	bestProducts, _ := s.billRepository.GetBestSellingProducts(10)
	bestProductIds := []uint{}
	for _, bp := range bestProducts {
		bestProductIds = append(bestProductIds, bp.ProductId)
	}

	var bills []respository.Bill_Header
	if limit > 0 {
		offset := (page - 1) * limit
		bills, err = s.billRepository.GetAllBill(filter, limit, offset, bestProductIds, sortOrder)
	} else {
		bills, err = s.billRepository.GetAllBill(filter, -1, -1, bestProductIds, sortOrder)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills: ", err)
	}

	// แปลง response
	var billResponses []Bill_HeaderResponse
	for _, p := range bills {
		billResponses = append(billResponses, Bill_HeaderResponse{
			Id:                     p.Id,
			Invoice:                p.Invoice,
			CreatedAt:              p.CreatedAt,
			UpdatedAt:              p.UpdatedAt,
			MemberId:               p.MemberId,
			MemberFullName:         p.Member.FullName,
			User_Id:                p.User_Id,
			UserFullName:           p.User.FullName,
			UserUsername:           p.User.Username,
			ProductId:              p.ProductId,
			ProductSku:             p.Product.Sku,
			ProductName:            p.Product.Name,
			ProductPrice:           p.Product.Price,
			ProductCategory:        p.Product.Category.Name,
			Extra_Percent:          p.Extra_Percent,
			Down_Percent:           p.Down_Percent,
			Installments_Month:     p.Installments_Month,
			Net_installment:        p.Net_installment,
			Total_Price:            p.Total_Price,
			Paid_Amount:            p.Paid_Amount,
			Remaining_Amount:       p.Remaining_Amount,
			Total_Installments:     p.Total_Installments,
			Paid_Installments:      p.Paid_Installments,
			Remaining_Installments: p.Remaining_Installments,
			Late_Day:               p.Late_Day,
			Fee_Amount:             p.Fee_Amount,
			Status:                 p.Status,
			Note:                   p.Note,

			Credit_Balance: p.Credit_Balance,
		})
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &PaginationResponseBill{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     limit > 0 && page < totalPages,
		HasPrev:     limit > 0 && page > 1,
		Limit:       limit,
		Header: BillHeaderSummary{
			PaidBillCount:   paid,
			UnpaidBillCount: unpaid,
		},
		Bills:     billResponses,
		SumPaid:   sumPaid, // ✅ รวมยอดจ่ายทั้งหมด (เฉพาะ status=1)
		FeeAmount: feeAmount,
	}, nil
}
func (s *billService) GetBillDetailById(id uint) (*Bill_DetailsResponse, error) {
	detail, err := s.billRepository.GetBillDetailById(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	response := Bill_DetailsResponse{

		Id:            detail.Id,
		Bill_HeaderId: detail.Bill_HeaderId,

		Installment_Price: detail.Installment_Price,

		Paid_Amount: detail.Paid_Amount,

		Payment_Date: detail.Payment_Date,

		UpdatedAt:  detail.UpdatedAt,
		Fee_Amount: detail.Fee_Amount,

		Status: detail.Status,

		Credit_Balance: detail.Credit_Balance,
		Payment_No:     detail.Payment_No,
	}
	return &response, nil

}
func (s *billService) GetBillById(id uint) (*Bill_HeaderResponse, error) {
	bill, err := s.billRepository.GetBillById(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	response := Bill_HeaderResponse{
		Id:                 bill.Id,
		Invoice:            bill.Invoice,
		CreatedAt:          bill.CreatedAt,
		UpdatedAt:          bill.UpdatedAt,
		MemberId:           bill.MemberId,
		MemberFullName:     bill.Member.FullName,
		User_Id:            bill.User_Id,
		UserFullName:       bill.User.FullName,
		UserUsername:       bill.User.Username,
		ProductId:          bill.ProductId,
		ProductSku:         bill.Product.Sku,
		ProductName:        bill.Product.Name,
		ProductPrice:       bill.Product.Price,
		ProductCategory:    bill.Product.Category.Name, // จาก relation Product.Category
		Extra_Percent:      bill.Extra_Percent,
		Down_Percent:       bill.Down_Percent,
		Installments_Month: bill.Installments_Month,
		Net_installment:    bill.Net_installment,
		Total_Price:        bill.Total_Price,
		Paid_Amount:        bill.Paid_Amount,

		Remaining_Amount: bill.Remaining_Amount,

		Total_Installments: bill.Total_Installments,

		Paid_Installments: bill.Paid_Installments,

		Remaining_Installments: bill.Remaining_Installments,

		Late_Day:   bill.Late_Day,
		Fee_Amount: bill.Fee_Amount,

		Status:         bill.Status,
		Note:           bill.Note,
		Credit_Balance: bill.Credit_Balance,
	}
	return &response, nil

}

func (s *billService) CreateInstallmentBill(request NewInstallmentBillHeader, installMentId uint) (*Bill_HeaderResponse_Installment, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	startDate := time.Now().In(loc)
	product, err := s.productRepository.GetProductByID(request.ProductId)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if request.MemberId == 0 {
		return nil, errors.New("member id is required")
	}

	if request.Loan_Amount >= product.Price {
		return nil, fmt.Errorf("ยอดยืม %.2f บาท มากกว่าราคาสินค้า %.2f บาท", request.Loan_Amount, product.Price)
	}

	// ค่าพื้นฐาน
	const fixedInterestPercent = 10.0
	loanAmount := request.Loan_Amount
	interAmount := request.Interest_Amount
	var interestAmount float64
	var totalPrice float64
	var netPrice float64
	var rentPrice float64
	var totalInstallments int
	var intsallDay int
	var totalPrice1 float64
	var cal1 float64
	var roundedRemaining float64

	// if installmentday.Day == 10 {
	// 	extraPercentResponse = 0 // หรือไม่ส่งค่ากลับเลยก็ได้
	// }

	if request.TermType == 1 && request.TermValue == 10 {
		// 🎯 กรณีผ่อน 10 วัน (รายวัน)
		// interestAmount = math.Round(loanAmount * fixedInterestPercent / 100)
		totalPrice = loanAmount
		totalPrice1 = loanAmount + interAmount
		cal1 = interAmount / fixedInterestPercent
		log.Print(cal1, "Killed")

		totalInstallments = 1
		intsallDay = 10
		rentPrice = totalPrice + cal1 // จ่ายครั้งเดียว
		fmt.Print(rentPrice, "rentPrice")
		if request.TermValue == 0 {
			return nil, errors.New("TermValue (จำนวนวันผ่อน) ต้องมากกว่า 0")
		}
		netPrice = totalPrice/10 + (cal1 / 10)
		log.Print(netPrice, "netPrice")
		request.Extra_Percent = 0
		interestAmount = interAmount
		roundedRemaining = math.Round(rentPrice)

	} else {
		// 🎯 กรณีผ่อนแบบรายเดือน
		// สูตร: (loan + (loan * extra_percent)) / month
		// totalInstallments = int(math.Ceil(float64(installmentday.Day) / 30.0))
		// if installmentday.Day%30 != 0 {
		// 	return nil, fmt.Errorf("installment day %d ไม่ถูกต้อง ต้องเป็นจำนวนวันที่หาร 30 ลงตัว เช่น 30, 60, 90", installmentday.Day)
		// }
		installmentday, err := s.installmentRepository.GetInstallmentById(installMentId)
		if err != nil {
			return nil, errors.New("installment not found")
		}
		if installmentday.Day == 0 {
			return nil, errors.New("Installment day must be greater than 0")
		}
		totalInstallments = installmentday.Day
		intsallDay = installmentday.Day
		percent := float64(request.Extra_Percent)
		interestAmount = math.Round(loanAmount * percent / 100)

		totalPrice1 = loanAmount + interestAmount
		fmt.Print()
		// totalInstallments = totalInstallments // แปลว่า จำนวนเดือนที่ต้องผ่อน
		rentPrice = math.Round(totalPrice1 / float64(totalInstallments))
		netPrice = math.Round(totalPrice1 / float64(totalInstallments))
		cal1 = interestAmount
		interAmount = cal1
		roundedRemaining = math.Round(totalPrice1)

	}

	// Generate invoice
	billNum, err := respository.GenerateHpc(s.billRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SKU: %v", err)
	}

	billHeader := &model.Bill_Header_Installment{
		Invoice:               billNum,
		MemberId:              request.MemberId,
		User_Id:               request.User_Id,
		ProductId:             request.ProductId,
		Extra_Percent:         request.Extra_Percent,
		Loan_Amount:           loanAmount,
		Interest_Amount:       cal1,
		Total_Price:           totalPrice1,
		Installment_Day:       intsallDay,
		Total_Installments:    totalInstallments,
		Net_installment:       netPrice,
		TermType:              request.TermType,
		Total_Interest_Amount: interAmount,
		Remaining_Amount:      roundedRemaining,
		LastRenewDate:         startDate,
		NextDueDate:           startDate.AddDate(0, 0, 10),

		Status: 1,
	}

	createdBill, err := s.billRepository.CreateInstallmentBill(billHeader)
	if err != nil {
		return nil, err
	}

	// ✅ Create installment details
	var details []model.Bill_Details_Installment
	for i := 1; i <= totalInstallments; i++ {
		var paymentDate time.Time

		if request.TermValue == 10 {
			// 🎯 จ่ายครั้งเดียว หลังจาก 10 วัน
			paymentDate = startDate.AddDate(0, 0, 10)
		} else {
			// 🎯 รายเดือน นับจากวันสร้างบิล
			paymentDate = startDate.AddDate(0, i, 0)
		}

		details = append(details, model.Bill_Details_Installment{
			Bill_Header_InstallmentId: createdBill.Id,
			Installment_Price:         rentPrice,
			Status:                    0,
			Payment_Date:              paymentDate,
			Payment_No:                fmt.Sprintf("%d", i),
		})
	}

	if err := s.billRepository.CreateInstallmentBillDetails(details); err != nil {
		return nil, err
	}

	resp := &Bill_HeaderResponse_Installment{
		Id:                    createdBill.Id,
		MemberId:              createdBill.MemberId,
		MemberFullName:        createdBill.Member.FullName,
		ProductId:             createdBill.ProductId,
		Total_Price:           createdBill.Total_Price,
		Net_installment:       createdBill.Net_installment,
		Total_Installments:    createdBill.Total_Installments,
		TermType:              createdBill.TermType,
		Total_Interest_Amount: createdBill.Total_Interest_Amount,
	}

	return resp, nil
}

func (s *billService) UpdateDailyInterest() error {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	today := time.Now().In(loc).Truncate(24 * time.Hour)

	bills, err := s.billRepository.GetAllUnpaid10DayBills()
	if err != nil {
		return err
	}

	for _, bill := range bills {
		if bill.Installment_Day != 10 || bill.Status != 1 {
			continue
		}
		if len(bill.BillDetailsInstallment) == 0 {
			log.Printf("⛔ ไม่มีงวดผ่อนในบิล %d", bill.Id)
			continue
		}

		var latestDetail *model.Bill_Details_Installment
		// (ส่วนการหา latestDetail เหมือนเดิม)
		for i := range bill.BillDetailsInstallment {
			d := &bill.BillDetailsInstallment[i]
			if d.Status == 0 {
				if latestDetail == nil || d.Payment_Date.After(latestDetail.Payment_Date) {
					latestDetail = d
				}
			}
		}
		if latestDetail == nil {
			// Fallback logic if no active installment is found
			for i := range bill.BillDetailsInstallment {
				d := &bill.BillDetailsInstallment[i]
				if latestDetail == nil || d.Payment_Date.After(latestDetail.Payment_Date) {
					latestDetail = d
				}
			}
		}

		if latestDetail == nil {
			log.Printf("⛔ ไม่มีข้อมูลงวดผ่อนในบิล %d", bill.Id)
			continue
		}

		dueDate := latestDetail.Payment_Date.In(loc).Truncate(24 * time.Hour)
		log.Printf("dueDate", dueDate)

		startInterestDate := dueDate.AddDate(0, 0, -10)
		log.Printf("startInterestDate", startInterestDate)

		if today.Before(startInterestDate) {
			log.Printf("⏩ ยังไม่ถึงวันเริ่มคิดดอก บิล %d (DueDate: %s, StartDate: %s)", bill.Id, dueDate.Format("2006-01-02"), startInterestDate.Format("2006-01-02"))
			continue
		}

		// --- 💡 [แก้ไข] ส่วนที่แก้ไขการนับวัน ---
		// คำนวณจำนวนวันที่ผ่านไป แล้วบวก 1 เพื่อให้นับ "วันนี้" รวมด้วยเสมอ
		daysLate := int(today.Sub(startInterestDate).Hours()/24) + 1
		log.Printf("daysLate", daysLate)
		if daysLate > 10 {
			daysLate = 10
		}
		// --- จบส่วนที่แก้ไข ---

		interestPerDay := float64(bill.Loan_Amount) * 0.10 / 10.0
		log.Printf("interestPerDay", interestPerDay)

		var daysAlreadyCharged int
		if interestPerDay > 0 {
			daysAlreadyCharged = int(math.Floor(bill.Interest_Amount / interestPerDay))
		}

		daysToCharge := daysLate - daysAlreadyCharged
		log.Printf("daysToCharge", daysToCharge)

		if daysToCharge <= 0 {
			log.Printf("✅ ข้ามบิล %d: ดอกเบี้ยครบแล้ว %d วัน (ควรคิด %d วัน)", bill.Id, daysAlreadyCharged, daysLate)
			continue
		}

		additional := interestPerDay * float64(daysToCharge)
		bill.Interest_Amount += additional

		log.Printf("✅ บวกดอกเพิ่ม %.2f (วันละ %.2f x %d วัน) | ดอกเบี้ยรวมตอนนี้ %.2f", additional, interestPerDay, daysToCharge, bill.Interest_Amount)

		total := float64(bill.Loan_Amount) + bill.Interest_Amount + bill.Fee_Amount
		bill.Net_installment = math.Round(total / 10.0)
		bill.Remaining_Amount = math.Round(total - float64(bill.Paid_Amount))

		latestDetail.Installment_Price = total
		if err := s.billRepository.UpdateInstallmentBillDetail1(latestDetail); err != nil {
			log.Printf("❌ ไม่สามารถอัปเดตผ่อนบิล %d: %v", latestDetail.Id, err)
			continue
		}

		if err := s.billRepository.UpdateBillInstallment(&bill); err != nil {
			log.Printf("❌ ไม่สามารถอัปเดตบิล %d: %v", bill.Id, err)
		} else {
			log.Printf("✅ อัปเดตดอกเบี้ย บิล %d | ดอกเบี้ยรวม %.2f | คิดดอกแล้ว %d วัน", bill.Id, bill.Interest_Amount, daysLate)
		}
	}

	return nil
}

func (s *billService) UpdateDailyInterestSingle(testDate ...time.Time) error {
	loc, _ := time.LoadLocation("Asia/Bangkok")

	// วันปัจจุบัน (รองรับ testDate สำหรับทดสอบ)
	var today time.Time
	if len(testDate) > 0 {
		t := testDate[0].In(loc)
		today = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	} else {
		now := time.Now().In(loc)
		today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	bills, err := s.billRepository.GetAllUnpaid10DayBills()
	if err != nil {
		return err
	}

	const termDays = 10

	for _, bill := range bills {
		if bill.Installment_Day != 10 || bill.Status != 1 {
			continue
		}

		if len(bill.BillDetailsInstallment) == 0 {
			// log.Printf("⛔ ไม่มีงวดผ่อนในบิล %d", bill.Id)
			continue
		}

		// --- Sort งวดผ่อนตาม Payment_Date descending ---
		sort.Slice(bill.BillDetailsInstallment, func(i, j int) bool {
			return bill.BillDetailsInstallment[i].Payment_Date.After(bill.BillDetailsInstallment[j].Payment_Date)
		})

		// --- หา latest interest installment ที่ยังไม่จ่าย ---
		var latestDetail *model.Bill_Details_Installment
		for i := range bill.BillDetailsInstallment {
			d := &bill.BillDetailsInstallment[i]
			if d.Is_Interest_Only && d.Status == 0 {
				latestDetail = d
				break
			}
		}
		if latestDetail == nil {
			// log.Printf("⛔ ไม่มีงวดผ่อนดอกคงค้างในบิล %d", bill.Id)
			continue
		}

		// Truncate เวลา
		dueDate := latestDetail.Payment_Date.In(loc)
		dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)

		// ✅ วันเริ่มรอบดอกเบี้ย
		startDate := dueDate.AddDate(0, 0, -termDays+1)
		// ✅ คำนวณจำนวนวันที่ผ่านมาในรอบ
		var daysPassed int
		if today.Before(startDate) {
			daysPassed = 0 // ยังไม่ถึงรอบ
		} else if today.After(dueDate) {
			daysPassed = termDays // ครบรอบเต็ม
		} else {
			daysPassed = int(today.Sub(startDate).Hours()/24) + 1
			if daysPassed > termDays {
				daysPassed = termDays
			}
		}

		// ✅ คำนวณดอกเบี้ยตามสัดส่วนวัน
		fullInterest := float64(bill.Loan_Amount) * 0.10
		interestPerDay := fullInterest / float64(termDays)
		expectedInterest := math.Round(interestPerDay * float64(daysPassed))

		// ✅ คำนวณ "ค่าปรับสะสม" (เช่น 20 บาท/วัน)
		dailyFee := 20.0
		expectedFee := float64(daysPassed) * dailyFee

		log.Printf("📅 Bill %d | start=%s | due=%s | today=%s | daysPassed=%d | interest=%.2f/%.2f | fee=%.2f",
			bill.Id,
			startDate.Format("2006-01-02"),
			dueDate.Format("2006-01-02"),
			today.Format("2006-01-02"),
			daysPassed,
			bill.Interest_Amount,
			expectedInterest,
			expectedFee,
		)

		// ✅ ถ้ายังไม่ถึง expectedInterest หรือ expectedFee ให้เพิ่ม
		needUpdate := false
		if math.Round(bill.Interest_Amount) < expectedInterest {
			additionalInterest := expectedInterest - bill.Interest_Amount
			bill.Interest_Amount += additionalInterest
			needUpdate = true
		}

		if math.Round(bill.Fee_Amount) < expectedFee {
			additionalFee := expectedFee - bill.Fee_Amount
			bill.Fee_Amount += additionalFee
			needUpdate = true
		}

		if !needUpdate {
			log.Printf("✅ ข้ามบิล %d: ดอกเบี้ยและค่าปรับครบ %d วัน", bill.Id, daysPassed)
			continue
		}

		// ✅ อัปเดตยอดรวม
		total := float64(bill.Loan_Amount) + bill.Interest_Amount
		bill.Net_installment = math.Round(total / float64(termDays))
		bill.Remaining_Amount = math.Round(total - float64(bill.Paid_Amount))
		latestDetail.Installment_Price = total

		fmt.Print("total", total)

		if err := s.billRepository.UpdateInstallmentBillDetail1(latestDetail); err != nil {
			// log.Printf("❌ ไม่สามารถอัปเดตผ่อนบิล %d: %v", latestDetail.Id, err)
			continue
		}

		if err := s.billRepository.UpdateBillInstallment(&bill); err != nil {
			// log.Printf("❌ ไม่สามารถอัปเดตบิล %d: %v", bill.Id, err)
		} else {
			// log.Printf("✅ อัปเดตดอกเบี้ย+ค่าปรับ บิล %d | ดอกเบี้ย %.2f | ค่าปรับ %.2f | สะสม %d วัน",
			// 	bill.Id, bill.Interest_Amount, bill.Fee_Amount, daysPassed)
		}

	}

	return nil
}


func (s *billService) PayPurchaseInstallment(billID uint, detailID uint, amount float64) ([]InstallmentPayResult, error) {
	results := []InstallmentPayResult{}

	// 1. ดึง Bill_Header
	bill, err := s.billRepository.GetInstallmentBillById(billID)
	if err != nil {
		return nil, errors.New("bill not found")
	}
	carryCredit := bill.Credit_Balance

	maxPayable := float64(bill.Total_Installments)*bill.Net_installment + bill.Fee_Amount + carryCredit
	remainingBill := maxPayable - float64(bill.Paid_Amount)
	if remainingBill < 0 {
		remainingBill = 0
	}

	installments, err := s.billRepository.GetUnpaidBillInstallments1(billID, detailID)
	if err != nil {
		return nil, errors.New("cannot get installments")
	}
	if len(installments) == 0 {
		return nil, errors.New("all installments already paid")
	}

	remainingAmount := amount
	// if remainingAmount > float64(bill.Remaining_Amount) {
	// 	return nil, fmt.Errorf("ยอดชำระ %.2f เกินยอดคงเหลือของบิล %.2f", remainingAmount, float64(bill.Remaining_Amount))
	// }
	fullPrice := bill.Total_Price
	dayPrice := float64(bill.Remaining_Amount) // ราคาที่อัพเดตตามวัน (เช่น 2040)
	if bill.Installment_Day == 10 {
		// ตรวจสอบว่าจ่ายยอดเท่ากับ fullPrice หรือ dayPrice เท่านั้น
		if amount != fullPrice && amount != dayPrice {
			return nil, fmt.Errorf("บิล 10 วัน ต้องจ่ายเต็มราคา %.2f หรือราคาตามวัน %.2f เท่านั้น", fullPrice, dayPrice)
		}
		if amount > bill.Remaining_Amount && amount != fullPrice {
			return nil, fmt.Errorf("ยอดชำระ1 %.2f เกินยอดคงเหลือของบิล1 %.2f", amount, bill.Remaining_Amount)
		}
	} else {
		// สำหรับบิลแบบอื่น ๆ ตรวจสอบยอดไม่เกิน Remaining_Amount
		if amount > bill.Remaining_Amount {
			return nil, fmt.Errorf("ยอดชำระ2 %.2f เกินยอดคงเหลือของบิล2 %.2f", amount, bill.Remaining_Amount)
		}
	}

	fmt.Print("666bill.Remaining_Amount+carryCredit//", bill.Remaining_Amount)
	// if remainingAmount > float64(bill.Remaining_Amount) {
	// 	return nil, fmt.Errorf("ยอดชำระ %.2f เกินยอดคงเหลือของบิล %.2f", remainingAmount, float64(bill.Remaining_Amount))
	// }

	for i := 0; i < len(installments) && (remainingAmount > 0 || carryCredit > 0); i++ {
		inst := &installments[i]
		unpaid := inst.Installment_Price - inst.Paid_Amount
		if unpaid <= 0 {
			continue
		}
		totalAvailable := remainingAmount + carryCredit
		if totalAvailable <= 0 {
			break
		}

		var result InstallmentPayResult
		result.InstallmentNo = i + 1

		// ถ้าเป็นบิล 10 วัน ให้จ่ายเฉพาะ Case A เท่านั้น
		// if bill.Installment_Day == 10 {
		// 	if totalAvailable == unpaid {

		// 		inst.Paid_Amount = inst.Installment_Price
		// 		inst.Status = 1
		// 		inst.UpdatedAt = time.Now()
		// 		if remainingAmount >= unpaid {
		// 			remainingAmount -= unpaid
		// 		} else {
		// 			carryCredit -= unpaid - remainingAmount
		// 			remainingAmount = 0
		// 		}
		// 		inst.Credit_Balance = 0
		// 		result.Case = "A"
		// 		result.Message = fmt.Sprintf("จ่ายพอดี งวด %d", i+1)
		// 		result.CreditLeft = 0
		// 		result.PaidAmount = inst.Paid_Amount
		// 	} else {
		// 		return nil, fmt.Errorf("บิล 10 วัน ต้องจ่ายพอดีเท่านั้น (Case A) งวด %d", i+1)
		// 	}
		// }
		if bill.Installment_Day == 10 {
			if totalAvailable == unpaid || totalAvailable == fullPrice {
				inst.Paid_Amount = inst.Installment_Price
				inst.Status = 1
				inst.UpdatedAt = time.Now()
				if remainingAmount >= totalAvailable {
					remainingAmount -= totalAvailable
				} else {
					carryCredit -= totalAvailable - remainingAmount
					remainingAmount = 0
				}
				inst.Credit_Balance = 0
				result.Case = "A"
				result.Message = fmt.Sprintf("จ่ายพอดี งวด %d", i+1)
				result.CreditLeft = carryCredit
				result.PaidAmount = inst.Paid_Amount
			}
			//  else {
			// 	return nil, fmt.Errorf("บิล 10 วัน ต้องจ่ายเต็มราคา %.2f หรือราคาตามวัน %.2f เท่านั้น (งวด %d)", fullPrice, unpaid, i+1)
			// }
		} else {
			fmt.Print("D")

			// Logic ปกติสำหรับบิลอื่น ๆ (ไม่ใช่ 10 วัน)
			if totalAvailable == unpaid {

				inst.Paid_Amount = inst.Installment_Price
				inst.Status = 1
				inst.UpdatedAt = time.Now()
				if remainingAmount >= unpaid {
					remainingAmount -= unpaid
				}
				inst.Credit_Balance = 0
				result.Case = "A"
				result.Message = fmt.Sprintf("จ่ายพอดี งวด %d", i+1)
				result.CreditLeft = 0
				result.PaidAmount = inst.Paid_Amount

			} else if totalAvailable > unpaid { // Case B
				fmt.Println("เข้า Case B: จ่ายเกิน สร้างเครดิตใหม่")

				inst.Paid_Amount = inst.Installment_Price
				inst.Status = 1
				inst.UpdatedAt = time.Now()
				over := totalAvailable - unpaid
				carryCredit = over
				remainingAmount = 0
				inst.Credit_Balance = carryCredit
				result.Case = "B"
				result.Message = fmt.Sprintf("จ่ายเกิน สร้างเครดิตใหม่ %.2f", carryCredit)
				result.CreditLeft = carryCredit
				result.PaidAmount = inst.Paid_Amount
				fmt.Print("B")

			} else if totalAvailable >= unpaid { // Case C
				fmt.Println("เข้า Case C: ใช้เครดิต ปิดงวด")

				inst.Paid_Amount = inst.Installment_Price
				inst.Status = 1
				inst.UpdatedAt = time.Now()
				need := unpaid - remainingAmount
				carryCredit -= need
				if carryCredit < 0 {
					carryCredit = 0
				}
				remainingAmount = 0
				inst.Credit_Balance = 0
				result.Case = "C"
				result.Message = fmt.Sprintf("ใช้เครดิต ปิดงวด เครดิตเหลือ %.2f", carryCredit)
				result.CreditLeft = carryCredit
				result.PaidAmount = inst.Paid_Amount
				fmt.Print("C")

			} else { // Case D
				if remainingAmount+carryCredit >= unpaid {
					inst.Paid_Amount = inst.Installment_Price
					inst.Status = 1
					inst.UpdatedAt = time.Now()
					over := remainingAmount + carryCredit - unpaid
					carryCredit = over
					remainingAmount = 0
					inst.Credit_Balance = 0
					result.Case = "D"
					result.Message = fmt.Sprintf("ปิดงวดด้วยเครดิตทั้งหมด เครดิตเหลือ %.2f", carryCredit)
					result.CreditLeft = carryCredit
					result.PaidAmount = inst.Paid_Amount
				}

			}
		}
		results = append(results, result)
	}

	// Update database
	if err := s.billRepository.UpdateInstallmentBillDetail(installments); err != nil {
		return nil, err
	}

	bill.Credit_Balance = carryCredit
	paidInstallments, err := s.billRepository.GetPaidInstallmentsBill1(billID)
	if err != nil {
		return nil, errors.New("cannot get paid installments from DB")
	}
	time.Sleep(100 * time.Millisecond)

	bill.Paid_Installments = len(paidInstallments)
	bill.Remaining_Installments = bill.Total_Installments - bill.Paid_Installments
	bill.Paid_Amount += int(amount - remainingAmount)
	bill.Remaining_Amount -= amount - remainingAmount
	if bill.Remaining_Amount <= 0 {
		bill.Remaining_Amount = 0
		bill.Status = 2
	}

	if err := s.billRepository.UpdateBillInstallment(bill); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *billService) GetInstallmentBillById(id uint) (*Bill_HeaderResponse_Installment, error) {
	bill, err := s.billRepository.GetInstallmentBillById(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	response := Bill_HeaderResponse_Installment{
		Id:              bill.Id,
		Invoice:         bill.Invoice,
		CreatedAt:       bill.CreatedAt,
		UpdatedAt:       bill.UpdatedAt,
		MemberId:        bill.MemberId,
		MemberFullName:  bill.Member.FullName,
		User_Id:         bill.User_Id,
		UserFullName:    bill.User.FullName,
		UserUsername:    bill.User.Username,
		ProductId:       bill.ProductId,
		ProductSku:      bill.Product.Sku,
		ProductName:     bill.Product.Name,
		ProductPrice:    bill.Product.Price,
		ProductCategory: bill.Product.Category.Name, // จาก relation Product.Category
		Extra_Percent:   bill.Extra_Percent,
		// Down_Percent:       bill.Down_Percent,
		// Installments_Month: bill.Installments_Month,
		Net_installment: bill.Net_installment,
		Total_Price:     bill.Total_Price,
		Paid_Amount:     bill.Paid_Amount,

		Remaining_Amount: bill.Remaining_Amount,

		Total_Installments: bill.Total_Installments,

		Paid_Installments: bill.Paid_Installments,

		Remaining_Installments: bill.Remaining_Installments,

		Late_Day:              bill.Late_Day,
		Fee_Amount:            bill.Fee_Amount,
		Status:                bill.Status,
		Note:                  bill.Note,
		TermType:              bill.TermType,
		Interest_Amount:       bill.Interest_Amount,
		Total_Interest_Amount: bill.Total_Interest_Amount,

		Loan_Amount:    bill.Loan_Amount,
		Credit_Balance: bill.Credit_Balance,
	}
	return &response, nil

}
func (s *billService) GetInstallmentBillDetailById(id uint) (*Bill_Details_Installment, error) {
	detail, err := s.billRepository.GetInstallmentBillDetailById(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	response := Bill_Details_Installment{
		Id:                        detail.Id,
		Bill_Header_InstallmentId: detail.Bill_Header_InstallmentId,

		Installment_Price: detail.Installment_Price,
		Paid_Amount:       detail.Paid_Amount,

		Payment_Date: detail.Payment_Date,
		UpdatedAt:    detail.UpdatedAt,

		Fee_Amount: detail.Fee_Amount,
		Status:     detail.Status,

		Credit_Balance: detail.Credit_Balance,
		Payment_No:     detail.Payment_No,
	}
	return &response, nil

}


func (s *billService) AutoApplyInstallementLateFees() error {
	start := time.Now()
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.FixedZone("Asia/Bangkok", 7*3600)
	}

	bills, err := s.billRepository.GetAllInstallmentUnpaidBills()
	if err != nil {
		return err
	}
	if len(bills) == 0 {
		log.Println("❌ ไม่มีบิลที่ยังไม่ได้ชำระ")
		return nil
	}

	fine, err := s.fineRepositoty.GetFineById(2)
	if err != nil {
		return err
	}
	dailyFee := fine.FineAmount
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	var updatedCount int64

	numJobs := len(bills)
	jobs := make(chan uint, numJobs)
	var wg sync.WaitGroup
	workers := 2

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			for billID := range jobs {
				func(bid uint) {
					defer wg.Done()
					bill, err := s.billRepository.GetInstallmentBillById(bid)
					if err != nil {
						log.Printf("worker %d: failed to load bill id %d: %v", workerID, bid, err)
						return
					}

					if bill.TermType == 1 && len(bill.BillDetailsInstallment) == 0 {
						return
					}

					installments, err := s.billRepository.GetUnpaidBillInstallments(bid)
					if err != nil || len(installments) == 0 {
						return
					}

					changed := false
					graceDays := 3
					for i := range installments {
						inst := &installments[i]
						if inst.Status != 0 {
							continue
						}

						dueDate := inst.Payment_Date.In(loc).Truncate(24*time.Hour).AddDate(0, 0, graceDays)

						if !today.After(dueDate) {
							continue
						}

						lateDays := int(today.Sub(dueDate).Hours() / 24)
						if lateDays <= 0 {
							continue
						}
						
						newFee := float64(lateDays) * dailyFee
						additionalFee := newFee - inst.Fee_Amount
						
						if additionalFee <= 0 {
							continue
						}

						// อัปเดตค่าปรับในงวดและบิล (ใน memory)
						inst.Fee_Amount = round2(newFee)
						inst.Installment_Price = round2(inst.Installment_Price + additionalFee)
						
						if lateDays > bill.Late_Day {
							bill.Late_Day = lateDays
						}

						bill.Fee_Amount = round2(bill.Fee_Amount + additionalFee)
						changed = true
					}

					if !changed {
						log.Printf("⏭ บิล %d ไม่มีการเปลี่ยนแปลงค่าปรับ", bill.Id)
						return
					}
					
					// ======================================================================
					// --- 💡 [แก้ไข] ส่วนที่ป้องกันการเขียนทับข้อมูล ---
					// ======================================================================

					// 1. ดึงข้อมูล bill ล่าสุดจาก DB อีกครั้ง!
					// เพื่อให้ได้ Interest_Amount ที่ UpdateDailyInterest เพิ่งบันทึกไป
					latestBill, err := s.billRepository.GetInstallmentBillById(bill.Id)
					if err != nil {
						log.Printf("❌ worker: ไม่สามารถดึงข้อมูลบิลล่าสุด id %d: %v", bill.Id, err)
						return
					}

					// 2. นำค่า Fee_Amount และ Late_Day ที่คำนวณใหม่ ไปใส่ใน latestBill
					latestBill.Fee_Amount = bill.Fee_Amount
					latestBill.Late_Day = bill.Late_Day

					// 3. คำนวณ Remaining_Amount ใหม่ด้วยข้อมูลที่ถูกต้องและครบถ้วนที่สุด
					latestBill.Remaining_Amount = math.Round(
						float64(latestBill.Loan_Amount) +
							latestBill.Interest_Amount + // <-- ใช้ดอกเบี้ยล่าสุดจาก DB
							latestBill.Fee_Amount -       // <-- ใช้ค่าปรับล่าสุดที่เพิ่งคำนวณ
							float64(latestBill.Paid_Amount),
					)
					log.Println("bill.Remaining_Amount (คำนวณใหม่)", latestBill.Remaining_Amount)

					// ======================================================================
					
					atomic.AddInt64(&updatedCount, 1)

					// บันทึกข้อมูล installment ที่มีการเปลี่ยนแปลงค่าปรับ
					if err := s.billRepository.UpdateInstallmentBillDetail(installments); err != nil {
						log.Printf("❌ ล้มเหลวอัปเดตงวดบิล %d: %v", latestBill.Id, err)
					}

					// บันทึก latestBill ที่มี Remaining_Amount ที่ถูกต้องลง DB
					if err := s.billRepository.UpdateBillFeeInstallment(latestBill); err != nil {
						log.Printf("❌ ล้มเหลวอัปเดตบิล %d: %v", latestBill.Id, err)
					} else {
						log.Printf("✅ อัปเดตค่าปรับบิล %d สำเร็จ", latestBill.Id)
					}

				}(billID)
			}
		}(w)
	}

	for _, b := range bills {
		wg.Add(1)
		jobs <- b.Id
	}
	close(jobs)
	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("AutoApplyInstallmentLateFees เสร็จใน %s, อัปเดต %d บิล", elapsed, updatedCount)
	return nil
}
func (s *billService) AddInstallmentExtraPayment(billID uint, installmentID uint, request UpdateAddExtraRequest_Installment) error {
	// 1. ดึง Bill_Header
	bill, err := s.billRepository.GetInstallmentBillById(billID)
	if err != nil {
		return errors.New("bill not found")
	}

	// 2. ดึง Bill_Detail ตาม installmentID
	inst, err := s.billRepository.GetInstallmentBillDetailById(installmentID)
	if err != nil {
		return errors.New("installment not found")
	}

	// 3. ตรวจสอบว่ายอดเงินที่จ่ายไม่เกินยอดที่เหลือของงวด
	// remainingInst := bill.Total_Price - inst.Paid_Amount
	// if request.Paid_Amount > remainingInst {
	// 	log.Printf("666Error: Payment exceeds remaining installment amount. Requested: %.2f, Remaining: %.2f", request.Paid_Amount, remainingInst)
	// 	return fmt.Errorf("66payment exceeds remaining amount of this installment: %.2f > %.2f", request.Paid_Amount, remainingInst)
	// }

	// 4. ตรวจสอบว่ายอดเงินไม่เกินยอดคงเหลือของบิล
	if request.Paid_Amount > bill.Remaining_Amount {
		return fmt.Errorf("payment exceeds remaining amount of the bill: %.2f > %d", request.Paid_Amount, bill.Remaining_Amount)
	}

	// 5. เพิ่มยอด Paid_Amount ของงวด
	inst.Paid_Amount += request.Paid_Amount

	// 6. อัปเดตยอดรวมหัวบิล
	bill.Paid_Amount += int(request.Paid_Amount)
	bill.Remaining_Amount -= request.Paid_Amount

	// 7. ปรับ Status ของงวดถ้าปิด
	if inst.Paid_Amount >= inst.Installment_Price {
		if request.Paid_Amount > bill.Remaining_Amount {
			inst.Status = 1

		} else {
			inst.Status = 1                                                         // ปิดงวด
			inst.Credit_Balance = round2(inst.Paid_Amount - inst.Installment_Price) // ถ้ามีเงินเกิน เก็บเป็นเครดิต
		}

	}

	// 8. อัปเดตจำนวนงวดที่ปิดในหัวบิล
	paidInstallments := 0
	for _, d := range bill.BillDetailsInstallment {
		if d.Status == 1 || d.Id == inst.Id && inst.Status == 1 {
			paidInstallments++
		}
	}
	bill.Paid_Installments = paidInstallments
	bill.Remaining_Installments = bill.Total_Installments - paidInstallments

	// 9. ถ้า bill.Remaining_Amount <= 0 → ปิดบิล
	if bill.Remaining_Amount <= 0 {
		bill.Status = 2 // ปิดบิล
	}

	// 10. Save DB
	if err := s.billRepository.UpdateBillInstallment(bill); err != nil {
		return err
	}

	if err := s.billRepository.UpdateInstallmentSingleBillDetail(inst); err != nil {
		return err
	}

	return nil
}

func (s *billService) GetAllInstallmentBill(
	invs []string,
	dateFrom, dateTo *time.Time,
	page, limit int,
	sortOrder int, // <-- เพิ่มตรงนี้
) (*PaginationResponseBillInstallment, error) {

	if page < 1 {
		page = 1
	}

	filter := respository.BillFilter{Invs: invs, DateFrom: dateFrom, DateTo: dateTo}

	total, err := s.billRepository.CountInstallmentBills(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count bills: ", err)
	}

	bestProducts, _ := s.billRepository.GetBestSellingInstallmentsProducts(10)
	bestProductIds := []uint{}
	for _, bp := range bestProducts {
		bestProductIds = append(bestProductIds, bp.ProductId)
	}
	paid, unpaid, err := s.billRepository.SumInstallmentPaidAndUnpaidCounts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize bills: ", err)
	}
	sumPaid, err := s.billRepository.SumPaidAmountByInstallmentStatus2(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}
	feeAmount, err := s.billRepository.SumFeeInstallmentByStatus2(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}
	var bills []model.Bill_Header_Installment
	if limit > 0 {
		offset := (page - 1) * limit
		bills, err = s.billRepository.GetInstallmentAllBill(filter, limit, offset, bestProductIds, sortOrder)
	} else {
		bills, err = s.billRepository.GetInstallmentAllBill(filter, -1, -1, bestProductIds, sortOrder)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills:", err)
	}

	// แปลง response
	var billResponses []Bill_HeaderResponse_Installment
	for _, p := range bills {
		billResponses = append(billResponses, Bill_HeaderResponse_Installment{
			Id:              p.Id,
			Invoice:         p.Invoice,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			MemberId:        p.MemberId,
			MemberFullName:  p.Member.FullName,
			User_Id:         p.User_Id,
			UserFullName:    p.User.FullName,
			UserUsername:    p.User.Username,
			ProductId:       p.ProductId,
			ProductSku:      p.Product.Sku,
			ProductName:     p.Product.Name,
			ProductPrice:    p.Product.Price,
			ProductCategory: p.Product.Category.Name,
			Extra_Percent:   p.Extra_Percent,
			// Down_Percent:           p.Down_Percent,
			// Installments_Month:     p.Installments_Month,
			Net_installment:        p.Net_installment,
			Total_Price:            p.Total_Price,
			Paid_Amount:            p.Paid_Amount,
			Remaining_Amount:       p.Remaining_Amount,
			Total_Installments:     p.Total_Installments,
			Paid_Installments:      p.Paid_Installments,
			Remaining_Installments: p.Remaining_Installments,
			Late_Day:               p.Late_Day,
			Fee_Amount:             p.Fee_Amount,
			Status:                 p.Status,
			Note:                   p.Note,
			Loan_Amount:            p.Loan_Amount,
			// Extra_Percent:          p.Extra_Percent,

			Total_Interest_Amount: p.Total_Interest_Amount,

			Credit_Balance: p.Credit_Balance,
		})
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &PaginationResponseBillInstallment{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     limit > 0 && page < totalPages,
		HasPrev:     limit > 0 && page > 1,
		Limit:       limit,
		Header: BillHeaderInstallmentSummary{
			PaidBillCount:   paid,
			UnpaidBillCount: unpaid,
		},
		Bills:     billResponses,
		SumPaid:   sumPaid, // ✅ รวมยอดจ่ายทั้งหมด (เฉพาะ status=1)
		FeeAmount: feeAmount,
	}, nil
}

func (s *billService) UpdateBill(id uint, request Update_Installment) (*Bill_HeaderResponse, error) {
	installment := &respository.Bill_Header{ // ✅ ต้องเป็น pointer
		Id:     id, // สำคัญ ต้อง set id ไม่งั้น update ไม่รู้ว่าจะ update record ไหน
		Status: request.Status,
		Note:   request.Note,
	}

	err := s.billRepository.UpdateBillStatus(installment)
	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลล่าสุดกลับมาแปลงเป็น response
	updatedInstallment, err := s.billRepository.GetBillById(id)
	if err != nil {
		return nil, err
	}

	return &Bill_HeaderResponse{
		Id:      updatedInstallment.Id,
		Invoice: updatedInstallment.Invoice,
		Status:  updatedInstallment.Status,
		Note:    updatedInstallment.Note,
	}, nil
}
func (s *billService) UpdateBill_Installment(id uint, request Update_Installment) (*Bill_HeaderResponse_Installment, error) {
	installment := &model.Bill_Header_Installment{ // ✅ ต้องเป็น pointer
		Id:     id, // สำคัญ ต้อง set id ไม่งั้น update ไม่รู้ว่าจะ update record ไหน
		Status: request.Status,
		Note:   request.Note,
	}

	err := s.billRepository.UpdateInstallmentBillStatus(installment)
	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลล่าสุดกลับมาแปลงเป็น response
	updatedInstallment, err := s.billRepository.GetInstallmentBillById(id)
	if err != nil {
		return nil, err
	}

	return &Bill_HeaderResponse_Installment{
		Id:      updatedInstallment.Id,
		Invoice: updatedInstallment.Invoice,
		Status:  updatedInstallment.Status,
		Note:    updatedInstallment.Note,
	}, nil
}

func (s *billService) GetAllBillUnpay(
	invs []string,
	dateFrom, dateTo *time.Time,
	page, limit int,
	sortOrder int, // <-- เพิ่มตรงนี้
	nameOrPhone []string,
) (*PaginationResponseBill, error) {

	if page < 1 {
		page = 1
	}

	filter := respository.BillFilter{Invs: invs, DateFrom: dateFrom, DateTo: dateTo, NameOrPhones: nameOrPhone}

	total, err := s.billRepository.CountUnpayBills(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count bills: ", err)
	}
	paid, unpaid, err := s.billRepository.SumPaidAndUnpaidCounts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize bills:", err)
	}
	sumPaid, err := s.billRepository.SumPaidAmountByStatus1(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}

	// bestProducts, _ := s.billRepository.GetBestSellingProducts(10)
	// bestProductIds := []uint{}
	// for _, bp := range bestProducts {
	// 	bestProductIds = append(bestProductIds, bp.ProductId)
	// }

	var bills []respository.Bill_Header
	if limit > 0 {
		offset := (page - 1) * limit
		bills, err = s.billRepository.GetUnPayAllBill(filter, limit, offset, sortOrder)
	} else {
		bills, err = s.billRepository.GetUnPayAllBill(filter, -1, -1, sortOrder)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills: ", err)
	}

	// แปลง response
	var billResponses []Bill_HeaderResponse
	for _, p := range bills {
		billResponses = append(billResponses, Bill_HeaderResponse{
			Id:                     p.Id,
			Invoice:                p.Invoice,
			CreatedAt:              p.CreatedAt,
			UpdatedAt:              p.UpdatedAt,
			MemberId:               p.MemberId,
			MemberFullName:         p.Member.FullName,
			User_Id:                p.User_Id,
			UserFullName:           p.User.FullName,
			UserUsername:           p.User.Username,
			ProductId:              p.ProductId,
			ProductSku:             p.Product.Sku,
			ProductName:            p.Product.Name,
			ProductPrice:           p.Product.Price,
			ProductCategory:        p.Product.Category.Name,
			Extra_Percent:          p.Extra_Percent,
			Down_Percent:           p.Down_Percent,
			Installments_Month:     p.Installments_Month,
			Net_installment:        p.Net_installment,
			Total_Price:            p.Total_Price,
			Paid_Amount:            p.Paid_Amount,
			Remaining_Amount:       p.Remaining_Amount,
			Total_Installments:     p.Total_Installments,
			Paid_Installments:      p.Paid_Installments,
			Remaining_Installments: p.Remaining_Installments,
			Late_Day:               p.Late_Day,
			Fee_Amount:             p.Fee_Amount,
			Status:                 p.Status,
			Note:                   p.Note,

			Credit_Balance: p.Credit_Balance,
		})
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &PaginationResponseBill{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     limit > 0 && page < totalPages,
		HasPrev:     limit > 0 && page > 1,
		Limit:       limit,
		Header: BillHeaderSummary{
			PaidBillCount:   paid,
			UnpaidBillCount: unpaid,
		},
		Bills:   billResponses,
		SumPaid: sumPaid, // ✅ รวมยอดจ่ายทั้งหมด (เฉพาะ status=1)

	}, nil
}

func (s *billService) GetAllInstallmentBillUnpay(
	invs []string,
	dateFrom, dateTo *time.Time,
	page, limit int,
	sortOrder int, // <-- เพิ่มตรงนี้
	nameOrPhone []string,

) (*PaginationResponseBillInstallment, error) {

	if page < 1 {
		page = 1
	}

	filter := respository.BillFilter{Invs: invs, DateFrom: dateFrom, DateTo: dateTo, NameOrPhones: nameOrPhone}

	total, err := s.billRepository.CountInstallmentBillsUnpay(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count bills: ", err)
	}

	// bestProducts, _ := s.billRepository.GetBestSellingInstallmentsProducts(10)
	// bestProductIds := []uint{}
	// for _, bp := range bestProducts {
	// 	bestProductIds = append(bestProductIds, bp.ProductId)
	// }
	paid, unpaid, err := s.billRepository.SumInstallmentPaidAndUnpaidCounts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize bills: ", err)
	}
	sumPaid, err := s.billRepository.SumPaidAmountByInstallmentStatus1(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to sum paid_amount:", err)
	}

	var bills []model.Bill_Header_Installment
	if limit > 0 {
		offset := (page - 1) * limit
		bills, err = s.billRepository.GetInstallmentAllBillUnpay(filter, limit, offset, sortOrder)
	} else {
		bills, err = s.billRepository.GetInstallmentAllBillUnpay(filter, -1, -1, sortOrder)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills: ", err)
	}

	// แปลง response
	var billResponses []Bill_HeaderResponse_Installment
	for _, p := range bills {
		billResponses = append(billResponses, Bill_HeaderResponse_Installment{
			Id:              p.Id,
			Invoice:         p.Invoice,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			MemberId:        p.MemberId,
			MemberFullName:  p.Member.FullName,
			User_Id:         p.User_Id,
			UserFullName:    p.User.FullName,
			UserUsername:    p.User.Username,
			ProductId:       p.ProductId,
			ProductSku:      p.Product.Sku,
			ProductName:     p.Product.Name,
			ProductPrice:    p.Product.Price,
			ProductCategory: p.Product.Category.Name,
			Extra_Percent:   p.Extra_Percent,
			// Down_Percent:           p.Down_Percent,
			// Installments_Month:     p.Installments_Month,
			Net_installment:        p.Net_installment,
			Total_Price:            p.Total_Price,
			Paid_Amount:            p.Paid_Amount,
			Remaining_Amount:       p.Remaining_Amount,
			Total_Installments:     p.Total_Installments,
			Paid_Installments:      p.Paid_Installments,
			Remaining_Installments: p.Remaining_Installments,
			Late_Day:               p.Late_Day,
			Fee_Amount:             p.Fee_Amount,
			Status:                 p.Status,
			Note:                   p.Note,
			TermType:               p.TermType,
			Credit_Balance:         p.Credit_Balance,
			Interest_Amount:        p.Interest_Amount,
			Total_Interest_Amount:  p.Total_Interest_Amount,
			Loan_Amount:            p.Loan_Amount,
		})
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &PaginationResponseBillInstallment{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     limit > 0 && page < totalPages,
		HasPrev:     limit > 0 && page > 1,
		Limit:       limit,
		Header: BillHeaderInstallmentSummary{
			PaidBillCount:   paid,
			UnpaidBillCount: unpaid,
		},
		Bills:   billResponses,
		SumPaid: float64(sumPaid), // ✅ รวมยอดจ่ายทั้งหมด (เฉพาะ status=1)

	}, nil
}

func (s *billService) GetBillDetailsByIdUnpaid(id uint) ([]Bill_DetailsResponse, error) {
	details, err := s.billRepository.GetUnpaidInstallments2(id)
	if err != nil {
		return nil, fmt.Errorf("bill details not found for ID: ", id)
	}

	var responses []Bill_DetailsResponse
	for _, detail := range details {
		responses = append(responses, Bill_DetailsResponse{
			Id:                detail.Id,
			Bill_HeaderId:     detail.Bill_HeaderId,
			Installment_Price: detail.Installment_Price,
			Paid_Amount:       detail.Paid_Amount,
			Payment_Date:      detail.Payment_Date,
			UpdatedAt:         detail.UpdatedAt,
			Fee_Amount:        detail.Fee_Amount,
			Status:            detail.Status,
			Credit_Balance:    detail.Credit_Balance,
			Payment_No:        detail.Payment_No,
		})
	}

	return responses, nil
}

// Credit_Balance float64 `json:"credit_balance"`
func (s *billService) GetInstallmentBillByIdUnpaid(id uint) ([]Bill_Details_Installment, error) {
	// ดึงรายการค่างวดที่ยังไม่ชำระทั้งหมด
	unpaidInstallments, err := s.billRepository.GetUnpaidBillInstallments2(id)
	if err != nil {
		return nil, fmt.Errorf("Bill with ID  not found", id)
	}

	// แปลงเป็น response
	var responses []Bill_Details_Installment
	for _, ins := range unpaidInstallments {
		responses = append(responses, Bill_Details_Installment{
			Id:                        ins.Id,
			Bill_Header_InstallmentId: ins.Bill_Header_InstallmentId,
			Installment_Price:         ins.Installment_Price,
			Paid_Amount:               ins.Paid_Amount,
			Payment_Date:              ins.Payment_Date,
			UpdatedAt:                 ins.UpdatedAt,
						CreatedAt:                 ins.CreatedAt,

			Fee_Amount:                ins.Fee_Amount,
			Status:                    ins.Status,
			Credit_Balance:            ins.Credit_Balance,
			Payment_No:                ins.Payment_No,
		})
	}

	return responses, nil
}

// Bill_DetailsResponse1
func (s *billService) GetDueTodayBillsWithInstallments(sortData string) (*BillResponseWrapperMap, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc).Truncate(24 * time.Hour)

	installments, err := s.billRepository.GetUnpaidInstallmentsByDate()
	if err != nil {
		return nil, err
	}

	var overdueBills, dueToday []Bill_DetailsResponse2

	for _, ins := range installments {
		paymentDate := ins.Payment_Date.In(loc).Truncate(24 * time.Hour)
		daysLeft := int(paymentDate.Sub(now).Hours() / 24)

		// สร้าง Bill_HeaderResponse1
		header := Bill_HeaderResponse2{
			Id:        ins.BillHeader.Id,
			Invoice:   ins.BillHeader.Invoice,
			CreatedAt: ins.BillHeader.CreatedAt,
			UpdatedAt: ins.BillHeader.UpdatedAt,
			DeletedAt: ins.BillHeader.DeletedAt,

			MemberId:       ins.BillHeader.MemberId,
			MemberFullName: ins.BillHeader.Member.FullName,
			MemberUser_id:  ins.BillHeader.Member.UserId,

			User_Id:      ins.BillHeader.User_Id,
			UserFullName: ins.BillHeader.User.FullName,
			UserUsername: ins.BillHeader.User.Username,

			ProductId:       ins.BillHeader.ProductId,
			ProductSku:      ins.BillHeader.Product.Sku,
			ProductName:     ins.BillHeader.Product.Name,
			ProductPrice:    ins.BillHeader.Product.Price,
			ProductCategory: ins.BillHeader.Product.Category.Name,

			Extra_Percent:      ins.BillHeader.Extra_Percent,
			Down_Percent:       ins.BillHeader.Down_Percent,
			Installments_Month: ins.BillHeader.Installments_Month,
			Net_installment:    ins.BillHeader.Net_installment,

			Total_Price:            ins.BillHeader.Total_Price,
			Paid_Amount:            ins.BillHeader.Paid_Amount,
			Remaining_Amount:       ins.BillHeader.Remaining_Amount,
			Total_Installments:     ins.BillHeader.Total_Installments,
			Paid_Installments:      ins.BillHeader.Paid_Installments,
			Remaining_Installments: ins.BillHeader.Remaining_Installments,

			Late_Day:   ins.BillHeader.Late_Day,
			Fee_Amount: ins.BillHeader.Fee_Amount,
			Status:     ins.BillHeader.Status,
			Note:       ins.BillHeader.Note,

			Credit_Balance: ins.BillHeader.Credit_Balance,
		}

		br := Bill_DetailsResponse2{
			Id:                ins.Id,
			Bill_HeaderId:     ins.Bill_HeaderId,
			Installment_Price: ins.Installment_Price,
			Paid_Amount:       ins.Paid_Amount,
			Payment_Date:      ins.Payment_Date,
			UpdatedAt:         ins.UpdatedAt,
			Fee_Amount:        ins.Fee_Amount,
			Status:            ins.Status,
			Credit_Balance:    ins.Credit_Balance,
			Payment_No:        ins.Payment_No,

			Bill_Header: []Bill_HeaderResponse2{header}, // ✅ wrap เป็น slice
		}

		if daysLeft < 0 {
			overdueBills = append(overdueBills, br)
		} else if daysLeft == 0 {
			dueToday = append(dueToday, br)
		}
	}

	// Sort overdueBills
	sort.Slice(overdueBills, func(i, j int) bool {
		return overdueBills[i].Payment_Date.Before(overdueBills[j].Payment_Date)
	})

	// Sort dueToday by sortData
	if strings.ToLower(sortData) == "desc" {
		sort.Slice(dueToday, func(i, j int) bool {
			return dueToday[i].Payment_Date.After(dueToday[j].Payment_Date)
		})
	} else {
		sort.Slice(dueToday, func(i, j int) bool {
			return dueToday[i].Payment_Date.Before(dueToday[j].Payment_Date)
		})
	}

	return &BillResponseWrapperMap{
		Results: map[string]interface{}{
			"overdue":   overdueBills,
			"due_today": dueToday,
		},
	}, nil
}

func (s *billService) GetDueTodayInstallmentBillsWithInstallments(sortData string) (*BillResponseWrapperMapInstall, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc).Truncate(24 * time.Hour)

	installments, err := s.billRepository.GetUnpaidInstallBillmentsByDate()
	if err != nil {
		return nil, err
	}

	var overdueBills, dueToday []Bill_Details_Installment2

	for _, ins := range installments {
		paymentDate := ins.Payment_Date.In(loc).Truncate(24 * time.Hour)
		daysLeft := int(paymentDate.Sub(now).Hours() / 24)
		header := Bill_HeaderResponse_Installment2{
			Id:        ins.Bill_Header_Installment.Id,
			Invoice:   ins.Bill_Header_Installment.Invoice,
			CreatedAt: ins.Bill_Header_Installment.CreatedAt,
			UpdatedAt: ins.Bill_Header_Installment.UpdatedAt,

			MemberId:       ins.Bill_Header_Installment.MemberId,
			MemberFullName: ins.Bill_Header_Installment.Member.FullName,
			MemberUser_id:  ins.Bill_Header_Installment.Member.UserId,

			User_Id:      ins.Bill_Header_Installment.User_Id,
			UserFullName: ins.Bill_Header_Installment.User.FullName,
			UserUsername: ins.Bill_Header_Installment.User.Username,

			ProductId:       ins.Bill_Header_Installment.ProductId,
			ProductSku:      ins.Bill_Header_Installment.Product.Sku,
			ProductName:     ins.Bill_Header_Installment.Product.Name,
			ProductPrice:    ins.Bill_Header_Installment.Product.Price,
			ProductCategory: ins.Bill_Header_Installment.Product.Category.Name,

			Extra_Percent:   ins.Bill_Header_Installment.Extra_Percent,
			Net_installment: ins.Bill_Header_Installment.Net_installment,

			Total_Price:            ins.Bill_Header_Installment.Total_Price,
			Paid_Amount:            ins.Bill_Header_Installment.Paid_Amount,
			Remaining_Amount:       ins.Bill_Header_Installment.Remaining_Amount,
			Total_Installments:     ins.Bill_Header_Installment.Total_Installments,
			Paid_Installments:      ins.Bill_Header_Installment.Paid_Installments,
			Remaining_Installments: ins.Bill_Header_Installment.Remaining_Installments,

			Late_Day:              ins.Bill_Header_Installment.Late_Day,
			Fee_Amount:            ins.Bill_Header_Installment.Fee_Amount,
			Status:                ins.Bill_Header_Installment.Status,
			Note:                  ins.Bill_Header_Installment.Note,
			Total_Interest_Amount: ins.Bill_Header_Installment.Total_Interest_Amount,
			Loan_Amount:           ins.Bill_Header_Installment.Loan_Amount,

			Credit_Balance: ins.Bill_Header_Installment.Credit_Balance,
		}

		br := Bill_Details_Installment2{
			Id:                        ins.Id,
			Bill_Header_InstallmentId: ins.Bill_Header_InstallmentId,
			Installment_Price:         ins.Installment_Price,
			Paid_Amount:               ins.Paid_Amount,
			Payment_Date:              ins.Payment_Date,
			UpdatedAt:                 ins.UpdatedAt,
			Fee_Amount:                ins.Fee_Amount,
			Status:                    ins.Status,
			Credit_Balance:            ins.Credit_Balance,
			Payment_No:                ins.Payment_No,

			Bill_Header: []Bill_HeaderResponse_Installment2{header}, // ✅ wrap เป็น slice

		}

		if daysLeft < 0 {
			overdueBills = append(overdueBills, br)
		} else if daysLeft == 0 {
			dueToday = append(dueToday, br)
		}
	}

	// Sort overdueBills by Payment_Date ASC
	sort.Slice(overdueBills, func(i, j int) bool {
		return overdueBills[i].Payment_Date.Before(overdueBills[j].Payment_Date)
	})

	// Sort dueToday by Payment_Date asc or desc
	if strings.ToLower(sortData) == "desc" {
		sort.Slice(dueToday, func(i, j int) bool {
			return dueToday[i].Payment_Date.After(dueToday[j].Payment_Date)
		})
	} else {
		sort.Slice(dueToday, func(i, j int) bool {
			return dueToday[i].Payment_Date.Before(dueToday[j].Payment_Date)
		})
	}

	return &BillResponseWrapperMapInstall{
		Results: map[string]interface{}{
			"overdue":   overdueBills,
			"due_today": dueToday,
		},
	}, nil
}
func (s *billService) GetUnpaidBillById(userId string) ([]Bill_HeaderResponse1, error) {
	details, err := s.billRepository.GetUnpaidBill(userId)
	if err != nil || len(details) == 0 {
		return nil, fmt.Errorf("bill not found")
	}

	// Group by bill_header_id
	groupedDetails := make(map[uint][]respository.Bill_Details1)
	headers := make(map[uint]respository.Bill_Header)

	for _, d := range details {
		headerID := d.Bill_HeaderId
		groupedDetails[headerID] = append(groupedDetails[headerID], d)
		headers[headerID] = d.BillHeader
	}

	// 🔥 เพิ่มการเรียง bill_header_id ให้ผลลัพธ์ไม่สลับลำดับ
	var sortedHeaderIDs []uint
	for headerID := range groupedDetails {
		sortedHeaderIDs = append(sortedHeaderIDs, headerID)
	}

	sort.Slice(sortedHeaderIDs, func(i, j int) bool {
		return sortedHeaderIDs[i] < sortedHeaderIDs[j]
	})

	// สร้างผลลัพธ์ตามลำดับที่เรียงแล้ว
	var result []Bill_HeaderResponse1

	for _, headerID := range sortedHeaderIDs {
		billDetails := groupedDetails[headerID]
		h := headers[headerID]

		var detailResponses []Bill_DetailsResponse1
		for _, d := range billDetails {
			detailResponses = append(detailResponses, Bill_DetailsResponse1{
				Id:                d.Id,
				Bill_HeaderId:     d.Bill_HeaderId,
				Installment_Price: d.Installment_Price,
				Paid_Amount:       d.Paid_Amount,
				Payment_Date:      d.Payment_Date,
				UpdatedAt:         d.UpdatedAt,
				Fee_Amount:        d.Fee_Amount,
				Status:            d.Status,

				Credit_Balance: d.Credit_Balance,
				Payment_No:     d.Payment_No,
			})
		}

		headerResponse := Bill_HeaderResponse1{
			Id:      h.Id,
			Invoice: h.Invoice,

			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
			DeletedAt: h.DeletedAt,

			MemberId:       h.MemberId,
			MemberFullName: h.Member.FullName,

			User_Id:      h.User_Id,
			UserFullName: h.User.FullName,
			UserUsername: h.User.Username,

			ProductId:       h.ProductId,
			ProductSku:      h.Product.Sku,
			ProductName:     h.Product.Name,
			ProductPrice:    h.Product.Price,
			ProductCategory: h.Product.Category.Name,

			Extra_Percent:      h.Extra_Percent,
			Down_Percent:       h.Down_Percent,
			Installments_Month: h.Installments_Month,
			Net_installment:    h.Net_installment,

			Total_Price:            h.Total_Price,
			Paid_Amount:            h.Paid_Amount,
			Remaining_Amount:       h.Remaining_Amount,
			Total_Installments:     h.Total_Installments,
			Paid_Installments:      h.Paid_Installments,
			Remaining_Installments: h.Remaining_Installments,
			Late_Day:               h.Late_Day,
			Fee_Amount:             h.Fee_Amount,
			Status:                 h.Status,
			Credit_Balance:         h.Credit_Balance,
			Note:                   h.Note,

			BillDetails: detailResponses,
		}

		result = append(result, headerResponse)
	}

	return result, nil
}

func (s *billService) GetUnpaidInstallmentBillById(userId string) ([]Bill_HeaderResponse_Installment1, error) {
	details, err := s.billRepository.GetUnpaidInstallmentBill(userId)
	if err != nil || len(details) == 0 {
		return nil, fmt.Errorf("bill not found")
	}

	// --- Group detail ตาม bill_header_installment_id ---
	groupedDetails := make(map[uint][]model.Bill_Details_Installment)
	headers := make(map[uint]model.Bill_Header_Installment)

	for _, d := range details {
		headerID := d.Bill_Header_InstallmentId
		groupedDetails[headerID] = append(groupedDetails[headerID], d)
		headers[headerID] = d.Bill_Header_Installment
	}
	var sortedHeaderIDs []uint
	for headerID := range groupedDetails {
		sortedHeaderIDs = append(sortedHeaderIDs, headerID)
	}

	sort.Slice(sortedHeaderIDs, func(i, j int) bool {
		return sortedHeaderIDs[i] < sortedHeaderIDs[j]
	})

	var response []Bill_HeaderResponse_Installment1

	for _, headerID := range sortedHeaderIDs {
		billDetails := groupedDetails[headerID]
		h := headers[headerID]

		// แปลง billDetails → Bill_Details_Installment1
		var detailResponses []Bill_Details_Installment1
		for _, d := range billDetails {
			detailResponses = append(detailResponses, Bill_Details_Installment1{
				Id:                        d.Id,
				Bill_Header_InstallmentId: d.Bill_Header_InstallmentId,
				Installment_Price:         d.Installment_Price,
				Paid_Amount:               d.Paid_Amount,
				Payment_Date:              d.Payment_Date,
				UpdatedAt:                 d.UpdatedAt,
				Fee_Amount:                d.Fee_Amount,
				Status:                    d.Status,
				Credit_Balance:            d.Credit_Balance,
				Payment_No:                d.Payment_No,

				Bill_Header: nil, // หลีกเลี่ยงการวน loop ซ้ำ
			})
		}

		headerResponse := Bill_HeaderResponse_Installment1{
			Id:      h.Id,
			Invoice: h.Invoice,

			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
			DeletedAt: h.DeletedAt.Time,

			MemberId:       h.MemberId,
			MemberFullName: h.Member.FullName,

			User_Id:      h.User_Id,
			UserFullName: h.User.FullName,
			UserUsername: h.User.Username,

			ProductId:       h.ProductId,
			ProductSku:      h.Product.Sku,
			ProductName:     h.Product.Name,
			ProductPrice:    h.Product.Price,
			ProductCategory: h.Product.Category.Name,

			Extra_Percent:    h.Extra_Percent,
			Net_installment:  h.Net_installment,
			Total_Price:      h.Total_Price,
			Paid_Amount:      h.Paid_Amount,
			Remaining_Amount: h.Remaining_Amount,

			Total_Installments:     h.Total_Installments,
			Paid_Installments:      h.Paid_Installments,
			Remaining_Installments: h.Remaining_Installments,

			Late_Day:              h.Late_Day,
			Fee_Amount:            h.Fee_Amount,
			Status:                h.Status,
			Credit_Balance:        h.Credit_Balance,
			Note:                  h.Note,
			TermType:              h.TermType,
			Interest_Amount:       h.Interest_Amount,
			Total_Interest_Amount: h.Total_Interest_Amount,
			Loan_Amount:           h.Loan_Amount,

			BillDetails: detailResponses,
		}

		response = append(response, headerResponse)
	}

	return response, nil
}

func (s *billService) GetpaidBillById(billID uint, detailID uint) ([]Bill_HeaderResponse1, error) {
	details, err := s.billRepository.GetpaidBill(billID, detailID)
	if err != nil || len(details) == 0 {
		return nil, fmt.Errorf("bill not found")
	}

	// Group by bill_header_id
	groupedDetails := make(map[uint][]respository.Bill_Details1)
	headers := make(map[uint]respository.Bill_Header)

	for _, d := range details {
		headerID := d.Bill_HeaderId
		groupedDetails[headerID] = append(groupedDetails[headerID], d)
		headers[headerID] = d.BillHeader
	}

	// 🔥 เพิ่มการเรียง bill_header_id ให้ผลลัพธ์ไม่สลับลำดับ
	var sortedHeaderIDs []uint
	for headerID := range groupedDetails {
		sortedHeaderIDs = append(sortedHeaderIDs, headerID)
	}

	sort.Slice(sortedHeaderIDs, func(i, j int) bool {
		return sortedHeaderIDs[i] < sortedHeaderIDs[j]
	})

	// สร้างผลลัพธ์ตามลำดับที่เรียงแล้ว
	var result []Bill_HeaderResponse1

	for _, headerID := range sortedHeaderIDs {
		billDetails := groupedDetails[headerID]
		h := headers[headerID]

		var detailResponses []Bill_DetailsResponse1
		for _, d := range billDetails {
			detailResponses = append(detailResponses, Bill_DetailsResponse1{
				Id:                d.Id,
				Bill_HeaderId:     d.Bill_HeaderId,
				Installment_Price: d.Installment_Price,
				Paid_Amount:       d.Paid_Amount,
				Payment_Date:      d.Payment_Date,
				UpdatedAt:         d.UpdatedAt,
				Fee_Amount:        d.Fee_Amount,
				Status:            d.Status,
				Credit_Balance:    d.Credit_Balance,
				Payment_No:        d.Payment_No,
			})
		}

		headerResponse := Bill_HeaderResponse1{
			Id:      h.Id,
			Invoice: h.Invoice,

			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
			DeletedAt: h.DeletedAt,

			MemberId:       h.MemberId,
			MemberFullName: h.Member.FullName,

			User_Id:      h.User_Id,
			UserFullName: h.User.FullName,
			UserUsername: h.User.Username,

			ProductId:       h.ProductId,
			ProductSku:      h.Product.Sku,
			ProductName:     h.Product.Name,
			ProductPrice:    h.Product.Price,
			ProductCategory: h.Product.Category.Name,

			Extra_Percent:      h.Extra_Percent,
			Down_Percent:       h.Down_Percent,
			Installments_Month: h.Installments_Month,
			Net_installment:    h.Net_installment,

			Total_Price:            h.Total_Price,
			Paid_Amount:            h.Paid_Amount,
			Remaining_Amount:       h.Remaining_Amount,
			Total_Installments:     h.Total_Installments,
			Paid_Installments:      h.Paid_Installments,
			Remaining_Installments: h.Remaining_Installments,
			Late_Day:               h.Late_Day,
			Fee_Amount:             h.Fee_Amount,
			Status:                 h.Status,
			Credit_Balance:         h.Credit_Balance,
			Note:                   h.Note,

			BillDetails: detailResponses,
		}

		result = append(result, headerResponse)
	}

	return result, nil
}

func (s *billService) GetpaidInstallmentBillById(billID uint, detailID uint) ([]Bill_HeaderResponse_Installment1, error) {
	details, err := s.billRepository.GetpaidInstallmentBill(billID, detailID)
	if err != nil || len(details) == 0 {
		return nil, fmt.Errorf("bill not found")
	}

	// --- Group detail ตาม bill_header_installment_id ---
	groupedDetails := make(map[uint][]model.Bill_Details_Installment)
	headers := make(map[uint]model.Bill_Header_Installment)

	for _, d := range details {
		headerID := d.Bill_Header_InstallmentId
		groupedDetails[headerID] = append(groupedDetails[headerID], d)
		headers[headerID] = d.Bill_Header_Installment
	}
	var sortedHeaderIDs []uint
	for headerID := range groupedDetails {
		sortedHeaderIDs = append(sortedHeaderIDs, headerID)
	}

	sort.Slice(sortedHeaderIDs, func(i, j int) bool {
		return sortedHeaderIDs[i] < sortedHeaderIDs[j]
	})

	var response []Bill_HeaderResponse_Installment1

	for _, headerID := range sortedHeaderIDs {
		billDetails := groupedDetails[headerID]
		h := headers[headerID]

		// แปลง billDetails → Bill_Details_Installment1
		var detailResponses []Bill_Details_Installment1
		for _, d := range billDetails {
			detailResponses = append(detailResponses, Bill_Details_Installment1{
				Id:                        d.Id,
				Bill_Header_InstallmentId: d.Bill_Header_InstallmentId,
				Installment_Price:         d.Installment_Price,
				Paid_Amount:               d.Paid_Amount,
				Payment_Date:              d.Payment_Date,
				UpdatedAt:                 d.UpdatedAt,
				Fee_Amount:                d.Fee_Amount,
				Status:                    d.Status,
				Credit_Balance:            d.Credit_Balance,
				Payment_No:                d.Payment_No,

				Bill_Header: nil, // หลีกเลี่ยงการวน loop ซ้ำ
			})
		}

		headerResponse := Bill_HeaderResponse_Installment1{
			Id:      h.Id,
			Invoice: h.Invoice,

			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
			DeletedAt: h.DeletedAt.Time,

			MemberId:       h.MemberId,
			MemberFullName: h.Member.FullName,

			User_Id:      h.User_Id,
			UserFullName: h.User.FullName,
			UserUsername: h.User.Username,

			ProductId:       h.ProductId,
			ProductSku:      h.Product.Sku,
			ProductName:     h.Product.Name,
			ProductPrice:    h.Product.Price,
			ProductCategory: h.Product.Category.Name,

			Extra_Percent:    h.Extra_Percent,
			Net_installment:  h.Net_installment,
			Total_Price:      h.Total_Price,
			Paid_Amount:      h.Paid_Amount,
			Remaining_Amount: h.Remaining_Amount,

			Total_Installments:     h.Total_Installments,
			Paid_Installments:      h.Paid_Installments,
			Remaining_Installments: h.Remaining_Installments,

			Late_Day:              h.Late_Day,
			Fee_Amount:            h.Fee_Amount,
			Status:                h.Status,
			Credit_Balance:        h.Credit_Balance,
			Note:                  h.Note,
			TermType:              h.TermType,
			Interest_Amount:       h.Interest_Amount,
			Total_Interest_Amount: h.Total_Interest_Amount,
			Loan_Amount:           h.Loan_Amount,

			BillDetails: detailResponses,
		}

		response = append(response, headerResponse)
	}

	return response, nil
}

func (s *billService) ApplyLateFeeToSingleBill(billID uint, today time.Time) error {
	log.Printf("🔍 [DEBUG] เริ่มทำ ApplyLateFeeToSingleBill billID: %d", billID)

	loc, _ := time.LoadLocation("Asia/Bangkok")
	todayDate := today.In(loc).Truncate(24 * time.Hour)

	fine, err := s.fineRepositoty.GetFineById(2)
	if err != nil {
		log.Printf("❌ ไม่สามารถดึงข้อมูลค่าปรับได้: %v", err)
		return err
	}
	feePerDay := fine.FineAmount
	log.Printf("ℹ️ อัตราค่าปรับต่อวัน: %.2f บาท", feePerDay)

	bill, err := s.billRepository.GetInstallmentBillById(billID)
	if err != nil {
		log.Printf("❌ ไม่พบข้อมูลบิล: %v", err)
		return err
	}

	installments, err := s.billRepository.GetUnpaidBillInstallments3(billID)
	if err != nil || len(installments) == 0 {
		log.Printf("⛔️ ไม่พบงวดที่ยังไม่ชำระ BillID: %d", billID)
		return nil
	}

	inst := &installments[len(installments)-1]
	if inst.Status != 0 {
		log.Printf("⏹ งวดล่าสุดถูกปิดแล้ว BillID: %d | Status: %d", billID, inst.Status)
		return nil
	}

	dueDate := inst.Payment_Date.In(loc).Truncate(24 * time.Hour)
	graceDays := 3
	log.Printf("📅 วันที่ครบกำหนดเดิม (DueDate): %s", dueDate.Format("2006-01-02"))
	log.Printf("📅 วันที่ตรวจสอบ (Today): %s", todayDate.Format("2006-01-02"))
	log.Printf("📎 จำนวนวันผ่อนผัน: %d วัน", graceDays)

	// คำนวณจำนวนวันที่สาย
	totalLateDays := int(todayDate.Sub(dueDate).Hours() / 24)
	log.Printf("⏱ รวมสายทั้งหมด: %d วัน", totalLateDays)

	if totalLateDays <= graceDays {
		log.Printf("✅ ยังอยู่ในช่วงผ่อนผัน (สาย %d วัน ≤ %d วัน)", totalLateDays, graceDays)
		return nil
	}

	// วันที่เริ่มคิดค่าปรับจริง ๆ
	startPenaltyDate := dueDate.AddDate(0, 0, graceDays)
	penaltyDays := totalLateDays - graceDays

	if penaltyDays <= 0 {
		log.Printf("❌ จำนวนวันที่ต้องปรับน้อยกว่าหรือเท่ากับ 0: %d", penaltyDays)
		return nil
	}

	log.Printf("📍 เริ่มคิดค่าปรับจากวันที่: %s", startPenaltyDate.Format("2006-01-02"))
	log.Printf("⏳ จำนวนวันที่ถูกปรับ: %d วัน (จาก %s ถึง %s)", penaltyDays,
		startPenaltyDate.Format("2006-01-02"),
		todayDate.Format("2006-01-02"))

	newTotalFee := float64(penaltyDays) * feePerDay
	additionalFee := newTotalFee - inst.Fee_Amount
	log.Printf("💰 ค่าปรับทั้งหมดที่ควรเป็น: %.2f บาท", newTotalFee)
	log.Printf("➕ ค่าปรับเพิ่มเติมจากเดิม: %.2f บาท (ค่าปรับเดิม: %.2f)", additionalFee, inst.Fee_Amount)

	if additionalFee <= 0 {
		log.Printf("⏭ ไม่มีการเพิ่มค่าปรับใหม่ เพราะค่าปรับเดิมครบถ้วนแล้ว")
		return nil
	}

	log.Printf("⚠️ คำนวณค่าปรับ BillID:%d | สายทั้งหมด %d วัน | ปรับ %d วัน | เพิ่มค่าปรับ %.2f",
		bill.Id, totalLateDays, penaltyDays, additionalFee)

	// อัปเดตค่าต่าง ๆ
	bill.Fee_Amount -= inst.Fee_Amount // ลบค่าปรับเก่าออกก่อน
	inst.Fee_Amount = round2(newTotalFee)
	bill.Fee_Amount += inst.Fee_Amount

	// คำนวณยอดคงเหลือใหม่
	bill.Remaining_Amount = math.Round(
		float64(bill.Loan_Amount) + bill.Interest_Amount + bill.Fee_Amount - float64(bill.Paid_Amount),
	)

	// อัปเดตวันที่สาย ถ้ามากกว่าเดิม
	if penaltyDays > bill.Late_Day {
		bill.Late_Day = penaltyDays
	}

	// Save ข้อมูล
	if err := s.billRepository.UpdateInstallmentBillDetail([]model.Bill_Details_Installment{*inst}); err != nil {
		log.Printf("❌ อัปเดต Installment ผิดพลาด: %v", err)
		return err
	}
	if err := s.billRepository.UpdateBillFeeInstallment(bill); err != nil {
		log.Printf("❌ อัปเดตค่าปรับใน Bill ผิดพลาด: %v", err)
		return err
	}

	log.Printf("✅ อัปเดตค่าปรับสำเร็จ | BillID:%d | ค่าปรับใหม่รวม %.2f บาท", bill.Id, inst.Fee_Amount)
	return nil
}

func (s *billService) RenewInterest(billID uint, payAmount float64, payDate time.Time) (*model.Bill_Header_Installment, error) {
	// loc, _ := time.LoadLocation("Asia/Bangkok")
	bill, err := s.billRepository.GetInstallmentBillById(billID)
	if err != nil {
		return nil, errors.New("ไม่พบบิล")
	}
	loc, _ := time.LoadLocation("Asia/Bangkok")
	// payDate = payDate.In(loc).Truncate(24 * time.Hour)
	payDate = time.Now().In(loc).Truncate(24 * time.Hour)
	lastRenew := bill.LastRenewDate.In(loc).Truncate(24 * time.Hour)
	nextDue := bill.NextDueDate.In(loc).Truncate(24 * time.Hour)
	fmt.Print(payDate, "paydate")

	if (payDate.Equal(lastRenew) || payDate.After(lastRenew)) && (payDate.Equal(nextDue) || payDate.Before(nextDue)) {
		log.Printf("✅ ยังอยู่ในรอบเดิม | BillID:%d", bill.Id)

		exists, err := s.billRepository.HasInterestInPeriod(bill.Id, lastRenew, nextDue)
		if err != nil {
			return nil, err
		}
		if exists {
			log.Printf("❌ มีรายการดอกเบี้ยแล้วในรอบนี้ | BillID:%d", bill.Id)
			return bill, nil
		}

		// =================================================================================
		// ✅ ขั้นตอนที่ 1: ปิดงวดเก่า และคำนวณยอดใหม่ตามสูตรที่คุณต้องการ
		// =================================================================================

		// 1.1) ปิดสถานะงวดเก่า (ทำให้เป็นเหมือนใบเสร็จ)
		allDetails, _ := s.billRepository.GetInstallmentDetailsByBillID(billID)
		for i := range allDetails {
			if allDetails[i].Status == 0 {
				allDetails[i].Status = 2
				// บันทึกว่ามีการจ่ายเงิน 200 ในงวดนี้ก่อนปิด
				// allDetails[i].Paid_Amount = payAmount
				_ = s.billRepository.UpdateInstallmentDetail(&allDetails[i])
			}
		}

		// 1.2) 💡 [แก้ไข] ใช้สูตรคำนวณที่คุณต้องการเพื่อให้ได้ผลลัพธ์ 2020
		const fixedInterestPercent = 10.0
		// cal1 จะได้เท่ากับ 200 / 10 = 20
		cal1 := payAmount // fixedInterestPercent
		log.Printf("cal1", cal1)
		log.Printf("payAmount", payAmount)

		// ยอดใหม่ = เงินต้น (2000) + cal1 (20) = 2020
		newInstallmentPrice := bill.Loan_Amount + cal1
		log.Printf("newInstallmentPrice", newInstallmentPrice)
		// =================================================================================
		// ✅ ขั้นตอนที่ 2: อัปเดต Bill Header และสร้างงวดใหม่
		// =================================================================================

		// 2.1) อัปเดต Bill Header ให้มีสถานะล่าสุด
		bill.Remaining_Amount = newInstallmentPrice // ยอดคงเหลือใหม่คือ 2020
		bill.Interest_Amount = cal1                 // ดอกเบี้ยสะสมสำหรับรอบใหม่คือ 20

		// 💡 [สำคัญ] เลื่อนวันครบกำหนดออกไปอีก 10 วัน!
		bill.LastRenewDate = payDate
		bill.NextDueDate = nextDue.AddDate(0, 0, 10)
		log.Printf("	bill.NextDueDate", bill.NextDueDate)

		// 2.2) สร้าง "งวดใหม่" ที่สะอาด สำหรับรอบบิลถัดไป
		newDetail := model.Bill_Details_Installment{
			Bill_Header_InstallmentId: bill.Id,
			Installment_Price:         bill.Remaining_Amount, // ใชัยอด 2020 ที่คำนวณได้
			Paid_Amount:               0,                     // งวดใหม่ยังไม่ได้จ่าย
			Payment_Date:              bill.NextDueDate,      // ใช้วันครบกำหนด "ใหม่"
			Status:                    0,                     // Active
			Is_Interest_Only:          true,
			Fee_Amount:                0,
			Payment_No:                fmt.Sprintf("INTEREST-%d-%s", bill.Id, payDate.Format("20060102150405")),
			UpdatedAt:                 payDate,
		}
		if err := s.billRepository.CreateInstallmentBillDetails([]model.Bill_Details_Installment{newDetail}); err != nil {
			return nil, errors.New("สร้างรายการงวดใหม่ไม่สำเร็จ")
		}

		// 2.3) บันทึก Bill Header ที่อัปเดตแล้ว
		// 💡 หมายเหตุ: เราจะยังไม่รีเซ็ต Interest_Amount ที่นี่ เพราะมันคือดอกเบี้ยของรอบใหม่
		if err := s.billRepository.UpdateBillInstallment(bill); err != nil {
			return nil, errors.New("อัปเดตข้อมูลบิลไม่สำเร็จ")
		}
		log.Printf("newDetail", newDetail.Payment_Date)
		log.Printf("lastday", bill.LastRenewDate)
		log.Printf("next", bill.NextDueDate)
		log.Printf("remaing_amount", bill.Remaining_Amount)
		log.Printf("interestment", bill.Interest_Amount)
		log.Printf("Fee_Amount", bill.Fee_Amount)
		log.Printf("Late_Day", bill.Late_Day)
		log.Printf("✅ จ่ายดอกเบี้ยและสร้างงวดใหม่สำเร็จ BillID:%d | New Remaining=%.2f",
			bill.Id, bill.Remaining_Amount)

		return bill, nil
	}

	if payDate.After(nextDue) {
		log.Printf("🔁 ต่อรอบใหม่ | BillID:%d | payDate=%s | oldNextDue=%s",
			bill.Id, payDate.Format("2006-01-02"), nextDue.Format("2006-01-02"))

		// =================================================================================
		// ✅ ขั้นตอนที่ 1: คำนวณดอกเบี้ยและค่าปรับที่ "ต้องชำระ" ทั้งหมดของงวดเก่า
		// =================================================================================
		allDetails, _ := s.billRepository.GetInstallmentDetailsByBillID(billID)

		billingCycleDays := 10
		totalLateDays := int(payDate.Sub(nextDue).Hours() / 24)
		log.Printf("🔄 เลยกำหนด %d วัน ", totalLateDays)

		var numberOfCycles int
		if totalLateDays > 0 {
			numberOfCycles = ((totalLateDays - 1) / billingCycleDays) + 1
		} else {
			numberOfCycles = 1
		}
		log.Printf("🔄 ตรวจสอบรอบบิล: เลยกำหนด %d วัน คิดเป็น %d รอบบิล", totalLateDays, numberOfCycles)

		// --- 1.2) คำนวณ "ยอดดอกเบี้ยที่ต้องชำระ" ของรอบเก่า ---
		interestPerCycle := bill.Loan_Amount * 0.10
		totalInterestForOldCycle := float64(numberOfCycles) * interestPerCycle
		log.Printf("✅ คำนวณดอกเบี้ยรอบเก่า: %d รอบ x %.2f บาท/รอบ | ดอกเบี้ยรวมที่ต้องชำระ %.2f",
			numberOfCycles, interestPerCycle, totalInterestForOldCycle)

		// --- 1.3) คำนวณ "ยอดค่าปรับที่ต้องชำระ" ของรอบเก่า ---
		var totalFeeForOldCycle float64
		graceDays := 3
		penaltyDays := totalLateDays - graceDays
		if penaltyDays > 0 {
			fine, err := s.fineRepositoty.GetFineById(2)
			if err != nil {
				return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลค่าปรับได้: %w", err)
			}
			feePerDay := fine.FineAmount
			totalFeeForOldCycle = float64(penaltyDays) * feePerDay

			log.Printf("⚠️ คำนวณค่าปรับรอบเก่า: ถูกปรับ %d วัน | ค่าปรับรวมที่ต้องชำระ %.2f",
				penaltyDays, totalFeeForOldCycle)
		}

		// --- 1.4) ตรวจสอบยอดชำระ ---
		totalDue := totalInterestForOldCycle + totalFeeForOldCycle
		if payAmount != totalDue {
			return nil, fmt.Errorf("ยอดชำระไม่ถูกต้อง: จ่าย %.2f แต่ยอดที่ต้องชำระคือ %.2f", payAmount, totalDue)
		}
		log.Printf("👍 BillID %d จ่ายถูกต้อง (ยอดชำระ %.2f)", bill.Id, payAmount)

		// =================================================================================
		// ✅ ขั้นตอนที่ 2: ปิดงวดเก่า และคำนวณสถานะของ "งวดใหม่"
		// =================================================================================

		// 2.1) ปิดสถานะงวดเก่า
		for i := range allDetails {
			if allDetails[i].Status == 0 {
				log.Printf("🧾 กำลังอัปเดตงวดเก่า (ID: %d) ให้เป็นใบเสร็จสรุปยอด...", allDetails[i].Id)

				allDetails[i].Status = 2 // ปิดสถานะ

				// ยอดรวมของธุรกรรมนี้ = เงินต้น + ยอดค้างชำระทั้งหมด
				allDetails[i].Installment_Price = bill.Loan_Amount + totalDue
				allDetails[i].Fee_Amount = totalFeeForOldCycle

				// เรียกใช้ฟังก์ชัน Repository ที่แก้ไขแล้ว
				if err := s.billRepository.UpdateInstallmentDetail(&allDetails[i]); err != nil {
					log.Printf("❌ เกิดข้อผิดพลาดในการอัปเดตงวดเก่า ID %d: %v", allDetails[i].Id, err)
					// เพื่อความปลอดภัย ควร return error ออกไป
					return nil, fmt.Errorf("ไม่สามารถอัปเดตงวดเก่า ID %d ได้: %w", allDetails[i].Id, err)
				} else {
					log.Printf("✅ อัปเดตงวดเก่า (ID: %d) สำเร็จ", allDetails[i].Id)
				}
			}
		}
		// 2.2) เงินต้นสำหรับรอบใหม่ คือเงินต้นเดิม
		newPrincipal := bill.Loan_Amount

		// 2.3) 💡 [แก้ไข] คำนวณดอกเบี้ยที่เกิดขึ้นแล้วสำหรับ "รอบใหม่" (Pro-rata)
		var interestForNewCycle float64 = 0
		// คำนวณหาว่าวันที่จ่าย อยู่ในวันที่เท่าไหร่ของ "รอบล่าสุด"
		daysIntoNewCycle := totalLateDays - ((numberOfCycles - 1) * billingCycleDays)

		if daysIntoNewCycle > 0 {
			interestPerDay := (newPrincipal * 0.10) / 10.0 // ดอกเบี้ยต่อวันจากเงินต้นใหม่
			log.Printf("🔄interestPerDay ", interestPerDay)

			interestForNewCycle = float64(daysIntoNewCycle) * interestPerDay
			log.Printf("🌀 คำนวณดอกเบี้ยล่วงหน้าสำหรับรอบใหม่: %d วัน | ดอกเบี้ย %.2f", daysIntoNewCycle, interestForNewCycle)
		}

		// 2.4) ยอดคงเหลือสุดท้าย (Remaining_Amount) คือ เงินต้นใหม่ + ดอกเบี้ยที่เกิดขึ้นแล้วในรอบใหม่
		// finalRemainingAmount := newPrincipal + interestForNewCycle
		// 		log.Printf("finalRemainingAmount ", finalRemainingAmount)
		for i := range allDetails {
			if allDetails[i].Status == 0 {
				allDetails[i].Status = 2
				allDetails[i].Fee_Amount = totalFeeForOldCycle
				allDetails[i].Installment_Price = newPrincipal + totalDue
				allDetails[i].Paid_Amount = 0

				log.Println("allDetails[i].Fee_Amount", allDetails[i].Fee_Amount)
				_ = s.billRepository.UpdateInstallmentDetail(&allDetails[i])
			}
		}

		// bill.Remaining_Amount = newPrincipal + (newPrincipal*0.10)/10.0
		bill.Remaining_Amount = newPrincipal + interestForNewCycle 

		log.Printf("bill.Remaining_Amount ", bill.Remaining_Amount)
		netInstallment := newPrincipal + interestForNewCycle
		bill.Net_installment = math.Round(netInstallment / float64(bill.Installment_Day))
		log.Printf("bill.Net_installment  ", bill.Net_installment)
		bill.Interest_Amount = interestForNewCycle
		log.Printf("bill.Interest_Amount  ", bill.Interest_Amount)

		log.Printf(" newPrincipal + totalDue", newPrincipal+totalDue)

		now := time.Now().In(loc)
		bill.LastRenewDate = now
		bill.NextDueDate = nextDue.AddDate(0, 0, numberOfCycles*billingCycleDays)
		log.Printf("📅 ตั้งวันครบกำหนดใหม่เป็น: %s", bill.NextDueDate.Format("2006-01-02"))

		bill.Fee_Amount = 0
		bill.Late_Day = 0
		log.Printf("📅 ค่าปรับ: %s", bill.Fee_Amount)
		// if err := s.UpdateDailyInterest(); err != nil {
		// 	return nil, errors.New("สร้างรายการงวดใหม่ไม่สำเร็จ")
		// }
		// 2.6) สร้างงวดใหม่เพื่อเก็บประวัติการต่ออายุ
		newDetail := model.Bill_Details_Installment{
			Bill_Header_InstallmentId: bill.Id,
			Installment_Price:         bill.Remaining_Amount, // ยอดรวมทั้งหมดของธุรกรรมนี้
			Paid_Amount:               0,
			Payment_Date:              bill.NextDueDate,
			Status:                    0,
			Is_Interest_Only:          true,
			Fee_Amount:                0 + bill.Fee_Amount,
			Payment_No:                fmt.Sprintf("RENEW-%s", payDate.Format("2006-01-02")),
			UpdatedAt:                 payDate,
		}

		// if err := s.UpdateDailyInterest(); err != nil {
		// 	return nil, errors.New("สร้างรายการงวดใหม่ไม่สำเร็จ")
		// }

		if err := s.billRepository.CreateInstallmentBillDetails([]model.Bill_Details_Installment{newDetail}); err != nil {
			return nil, errors.New("สร้างรายการงวดใหม่ไม่สำเร็จ")
		}

		// =================================================================================
		// ✅ ขั้นตอนที่ 3: ✨ ตั้งค่า Bill Header สำหรับ "สถานะปัจจุบันของรอบใหม่" ✨
		// =================================================================================
		// bill.Interest_Amount = interestForNewCycle // ดอกเบี้ยที่เกิดขึ้นแล้วในรอบใหม่
		// bill.Interest_Amount = math.Round((netInstallment / float64(bill.Installment_Day)) / 10)

		bill.Fee_Amount = 0
		bill.Late_Day = 0
		log.Printf("newDetail", newDetail.Payment_Date)

		if err := s.billRepository.UpdateBillInstallment(bill); err != nil {
			return nil, errors.New("อัปเดตข้อมูลบิลไม่สำเร็จ")
		}

		log.Printf("✅ ต่ออายุและเริ่มรอบใหม่สำเร็จ BillID:%d | New Remaining=%.2f",
			bill.Id, bill.Remaining_Amount)
	}
	return bill, nil
}
