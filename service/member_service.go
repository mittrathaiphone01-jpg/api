package service

import (
	"fmt"
	"rrmobile/respository"
)

type memberService struct {
	memberRepository respository.MemberRepository
}

func NewMemberService(memberRepository respository.MemberRepository) MemberService {
	return &memberService{memberRepository: memberRepository}
}


func (s *memberService) GetAllMembers(fullname, user_id []string, page, limit int) (*PaginationResponseMember, error) {
	if page < 1 {
		page = 1
	}

	offset := 0
	if limit > 0 {
		offset = (page - 1) * limit
	}

	filter := respository.MemberFilter{
		FullName: fullname,
		UserId:   user_id,
	}

	total, err := s.memberRepository.CountMembers(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: ", err)
	}

	products, err := s.memberRepository.GetMembers(filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: ", err)
	}

	var memberResponses []MemberResponse
	for _, p := range products {
		memberResponses = append(memberResponses, MemberResponse{
			Id:       p.Id,
			FullName: p.FullName,
			Tel:      p.Tel,
		})
	}

	// ถ้าไม่ limit → มีแค่หน้าเดียว
	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &PaginationResponseMember{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     limit > 0 && page < totalPages,
		HasPrev:     limit > 0 && page > 1,
		Limit:       limit,
		Member:      memberResponses,
	}, nil
}

func (s *memberService) GetMemberById(id uint) (*MemberResponse, error) {
	member, err := s.memberRepository.GetMemberById(id)
	if err != nil {
		return nil, err
	}
	return &MemberResponse{
		Id:       member.Id,
		FullName: member.FullName,
		Tel:      member.Tel,
	}, nil
}
func (s *memberService) CreateMember(req NewMemberRequest) (*MemberResponse, error) {
	member := respository.Member{
		FullName: req.FullName,
		Tel:      req.Tel,
		UserId:   req.UserId,
	}
	createdMember, err := s.memberRepository.AddMember(member)
	if err != nil {
		return nil, err
	}
	return &MemberResponse{
		Id:       createdMember.Id,
		FullName: createdMember.FullName,
		Tel:      createdMember.Tel,
	}, nil
}

func (s *memberService) EditMember(id uint, req UpdateMemberRequest) (*MemberResponse, error) {
	// ดึงข้อมูลเดิม
	member, err := s.memberRepository.GetMemberById(id)
	if err != nil {
		return nil, err
	}

	// อัปเดตเฉพาะ field ที่ถูกส่งมา
	if req.FullName != "" && req.FullName != member.FullName {
		member.FullName = req.FullName
	}

	if req.Tel != nil && *req.Tel != member.Tel {
		member.Tel = *req.Tel
	}

	if req.UserId != nil && *req.UserId != member.UserId {
		member.UserId = *req.UserId
	}

	updatedMember, err := s.memberRepository.UpdateMember(*member)
	if err != nil {
		return nil, err
	}

	return &MemberResponse{
		Id:       updatedMember.Id,
		FullName: updatedMember.FullName,
		Tel:      updatedMember.Tel,
	}, nil
}

func (s *memberService) DeleteMember(id uint) error {
	return s.memberRepository.DeleteMember(id)
}

func (s *memberService) GetMemberByUserId(userID string) (*MemberResponse, error) {
	member, err := s.memberRepository.CheckUserId(userID)
	if err != nil {
		return nil, err
	}

	return &MemberResponse{
		Id:       member.Id,
		FullName: member.FullName,
		Tel:      member.Tel,
	}, nil
}

func (s *memberService) UpdateMemberByTel(req UpdateMemberRequest1) (*MemberResponse, error) {
	// ต้องส่งเบอร์โทรมา เพื่อหา member
	if req.Tel == nil || *req.Tel == "" {
		return nil, fmt.Errorf("tel is required")
	}

	// ดึง member ตามเบอร์
	member, err := s.memberRepository.FindByTel(*req.Tel)
	if err != nil {
		return nil, fmt.Errorf("กรุณาให้พนักงาน บันทึกในระบบก่อนถึงจะชำระเงินได้")
	}

	// ถ้ามี user_id อยู่แล้ว และพยายามแก้ไข (ไม่ใช่ลบหรือว่าง)
	if member.UserId != "" && (req.UserId != nil && *req.UserId != "") {
		return nil, fmt.Errorf("มีข้อมูลแล้ว สามารถชำระได้เลย")
	}

	// ถ้าจะใส่ user_id ใหม่
	if req.UserId != nil {
		if *req.UserId == "" {
			// บังคับไม่ให้ user_id เป็นค่าว่าง
			return nil, fmt.Errorf("user_id cannot be empty")
		}
		// เช็คว่าซ้ำไหม
		if err := s.memberRepository.IsUserIdExists(*req.UserId); err != nil {
			return nil, err
		}

		member.UserId = *req.UserId // อัปเดต user_id
	}

	updated, err := s.memberRepository.UpdateMember(*member)
	if err != nil {
		return nil, err
	}

	return &MemberResponse{
		Id:       updated.Id,
		FullName: updated.FullName,
		Tel:      updated.Tel,
	}, nil
}
