package config

import (
	"fmt"
	"log"
	"rrmobile/model"

	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {

	hostDB := viper.GetString("DB_HOST")
	portDB := viper.GetString("DB_PORT")
	dbName := viper.GetString("DATABASE_NAME")
	user := viper.GetString("DB_USER")
	passwordDB := viper.GetString("DB_PASSWORD")
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		hostDB, portDB, user, passwordDB, dbName,
	)
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error getting database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		log.Fatalf("Error setting database connection pool: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	db.AutoMigrate(
		&model.Users{},
		&model.Role{},
		&model.RefreshToken{},
		&model.ProductCategory{},
		&model.Product{},
		&model.ProductImage{},
		&model.Rules{},
		&model.Installment{},
		&model.Fine_System{},
		&model.Fine_System_Category{},
		&model.Member{},
		&model.Bill_Header{},
		&model.Bill_Details{},
		&model.AccessToken{},
		&model.Bill_Header_Installment{},
		&model.Bill_Details_Installment{},
	)

	return db

}
