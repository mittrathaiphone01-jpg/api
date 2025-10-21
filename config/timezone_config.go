package config

import (
	"fmt"
	"time"
)

func InitTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
}
func ConvertToThaiTime(utcTime time.Time) (time.Time, error) {
	// โหลด Location ของโซนเวลาไทย
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading location: %v", err)
	}

	// แปลงเวลา UTC เป็นเวลามาตรฐานของไทย
	thaiTime := utcTime.In(location)
	return thaiTime, nil
}
