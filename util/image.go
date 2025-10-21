package util

import (
	"fmt"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

func GenerateFileName(original string) string {
	ext := filepath.Ext(original)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}
func ResizeImage(srcPath, dstPath string, maxWidth, maxHeight int) error {
	// เปิดไฟล์ภาพ
	img, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// ปรับขนาดภาพ โดยรักษาสัดส่วน
	resized := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	// บันทึกภาพใหม่ พร้อมคุณภาพสูง (เช่น 95%)
	err = imaging.Save(resized, dstPath, imaging.JPEGQuality(95))
	if err != nil {
		return err
	}

	return nil
}
