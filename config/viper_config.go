package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func ReloadEnv() error {
	// รีเซ็ตค่าใน viper ให้เหมือนเริ่มใหม่
	viper.Reset()
	// โหลดไฟล์ .env อีกครั้ง
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		return err
	}
	viper.AutomaticEnv()
	return nil
}
