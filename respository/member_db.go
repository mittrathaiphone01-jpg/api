package respository

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type memberRepositoryDB struct {
	db *gorm.DB
}

func NewMemberRepositoryDB(db *gorm.DB) MemberRepository {
	return &memberRepositoryDB{db: db}
}

func (r *memberRepositoryDB) GetMembers(filter MemberFilter, limit, offset int) ([]Member, error) {
	var members []Member
	query := r.db.Model(&Member{})

	// กรอง user_id
	var cleanUserIds []string
	for _, Usid := range filter.UserId {
		if Usid != "" {
			cleanUserIds = append(cleanUserIds, Usid)
		}
	}
	if len(cleanUserIds) > 0 {
		query = query.Where("user_id IN ?", cleanUserIds)
	}

	// กรอง full_name ด้วย ILIKE
	var likeFullNames []string
	for _, fullname := range filter.FullName {
		if fullname != "" {
			likeFullNames = append(likeFullNames, "%"+fullname+"%")
		}
	}
	if len(likeFullNames) > 0 {
		query = query.Where("full_name ILIKE ANY(?)", pq.Array(likeFullNames))
	}

	query = query.Order("id DESC")

	// ✅ เงื่อนไขสำคัญ: ถ้า limit > 0 ค่อยใช้ Limit/Offset
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&members).Error
	return members, err
}

func (r *memberRepositoryDB) CountMembers(filter MemberFilter) (int64, error) {
	var count int64
	query := r.db.Model(&Member{})

	// ✅ Filter UserId (เหมือน GetMembers)
	var cleanUserIds []string
	for _, id := range filter.UserId {
		if id != "" {
			cleanUserIds = append(cleanUserIds, id)
		}
	}
	if len(cleanUserIds) > 0 {
		query = query.Where("user_id IN ?", cleanUserIds)
	}

	// ✅ Filter FullName (ใช้ ILIKE ANY เหมือนกัน)
	var likeFullNames []string
	for _, fullname := range filter.FullName {
		if fullname != "" {
			likeFullNames = append(likeFullNames, "%"+fullname+"%")
		}
	}
	if len(likeFullNames) > 0 {
		query = query.Where("full_name ILIKE ANY(?)", pq.Array(likeFullNames))
	}

	// ✅ Count
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *memberRepositoryDB) GetMemberById(id uint) (*Member, error) {
	var member Member
	if err := r.db.First(&member, id).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepositoryDB) AddMember(member Member) (*Member, error) {
	var count int64

	// สร้าง query เริ่มต้น เช็ค full_name และ tel
	query := r.db.Model(&Member{}).
		Where("full_name = ? OR tel = ?", member.FullName, member.Tel)

	// ถ้า user_id ไม่ว่าง ให้เพิ่มเงื่อนไขเช็ค user_id ด้วย (OR)
	if member.UserId != "" {
		query = query.Or("user_id = ?", member.UserId)
	}

	// นับจำนวน record ที่ซ้ำ
	err := query.Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, fmt.Errorf("มีสมาชิกที่ใช้ชื่อ, เบอร์โทร หรือ user_id นี้แล้ว")
	}

	// สร้าง member ใหม่
	if err := r.db.Create(&member).Error; err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *memberRepositoryDB) isDuplicate(field string, value interface{}, id uint) error {
	var count int64
	if err := r.db.Model(&Member{}).Where(field+" = ? AND id != ?", value, id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("already exists", field)
	}
	return nil
}

func (r *memberRepositoryDB) UpdateMember(member Member) (*Member, error) {
	var existing Member
	if err := r.db.First(&existing, member.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Member not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}

	if member.FullName != "" && member.FullName != existing.FullName {
		if err := r.isDuplicate("full_name", member.FullName, member.Id); err != nil {
			return nil, err
		}
		updates["full_name"] = member.FullName
	}

	if member.Tel != existing.Tel {
		// ถ้าไม่ใช่ค่าว่าง ต้องตรวจสอบซ้ำ
		if member.Tel != "" {
			if err := r.isDuplicate("tel", member.Tel, member.Id); err != nil {
				return nil, err
			}
		}
		updates["tel"] = member.Tel
	}

	
	if member.UserId != existing.UserId {
		if member.UserId == "" {
			// ✅ อนุญาตให้ล้าง user_id ได้เลย ไม่ต้องเช็ค tel
			updates["user_id"] = ""
		} else {
			// ตรวจสอบว่า user_id ซ้ำหรือไม่
			if err := r.isDuplicate("user_id", member.UserId, member.Id); err != nil {
				return nil, err
			}
			updates["user_id"] = member.UserId
		}
	}

	if len(updates) == 0 {
		return &existing, nil
	}

	if err := r.db.Model(&Member{}).Where("id = ?", member.Id).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := r.db.First(&existing, member.Id).Error; err != nil {
		return nil, err
	}

	return &existing, nil
}

func (r *memberRepositoryDB) DeleteMember(id uint) error {
	if err := r.db.Delete(&Member{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *memberRepositoryDB) CheckUserId(userID string) (*Member, error) {
	var member Member
	if err := r.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบข้อมูลที่ต้องการ")
		}
		return nil, err
	}
	return &member, nil
}

func (r *memberRepositoryDB) FindByTel(tel string) (*Member, error) {
	var m Member
	if err := r.db.Where("tel = ?", tel).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("member with tel not found", tel)
		}
		return nil, err
	}
	return &m, nil
}

func (r *memberRepositoryDB) IsUserIdExists(userId string) error {
	var count int64
	if err := r.db.Model(&Member{}).Where("user_id = ?  ", userId).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("มีข้อมูลในระบบแล้ว สามารถชำระเงินได้เลย")
	}
	return nil
}
