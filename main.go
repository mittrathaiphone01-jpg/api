package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"rrmobile/backup"
	"rrmobile/config"
	"rrmobile/handler"
	"rrmobile/path"
	"rrmobile/respository"
	"rrmobile/service"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/middleware/helmet"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("ไม่สามารถโหลด .env ได้")
	}
	config.InitTimeZone()
	err = config.ReloadEnv()
	if err != nil {
		log.Fatal("ไม่สามารถโหลด environment variables ได้:", err)
	}
	if viper.GetString("SECRET_KEY") == "" {
		log.Fatal("SECRET_KEY is not set in environment variables")
	}

	db := config.InitDatabase()
	if db == nil {
		log.Fatal("ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
	}

	// app := fiber.New()
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		// BodyLimit:   50 * 1024 * 1024, // 50 MB

	})
	// app := fiber.New(fiber.Config{
	// 	// ตั้งค่า BodyLimit ให้รองรับไฟล์ใหญ่ เช่น 50MB
	// 	BodyLimit: 50 * 1024 * 1024, // 50 MB
	// })

	app.Use(fiberRecover.New())
	viper.SetConfigFile(".env") // หรือ config.yaml
	viper.ReadInConfig()

	// --- 💡 เพิ่มโค้ดส่วนนี้เพื่อพิมพ์ค่าออกมาดู ---
	corsOrigins := viper.GetString("CORS")
	log.Println("=============================================")
	log.Printf("🔍 Checking CORS Config: Read value for 'CORS' is -> '%s'", corsOrigins)
	log.Println("=============================================")
	app.Use(cors.New(cors.Config{
		// AllowOrigins: "*", // ต้องใส่ origin ที่เจาะจง
		AllowOrigins:     corsOrigins,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))
	app.Use(pprof.New())
	app.Use(helmet.New())
	app.Static("/uploads", "../uploads", fiber.Static{
		Compress: true,
	})

	app.Get("/image", func(c *fiber.Ctx) error {
		tokenString := c.Query("token")
		if tokenString == "" {
			return c.Status(400).SendString("Missing token")
		}

		secret := strings.TrimSpace(viper.GetString("SECRET_KEY"))
		if secret == "" {
			return c.Status(500).SendString("SECRET_KEY not configured")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil {
			return c.Status(401).SendString("Token error: " + err.Error())
		}
		if !token.Valid {
			return c.Status(401).SendString("Token is not valid")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(400).SendString("Invalid token claims")
		}

		// ดึง filename จาก token
		filename, ok := claims["filename"].(string)
		if !ok {
			return c.Status(400).SendString("Invalid token data: missing filename")
		}

		// ป้องกัน path traversal
		filePath := filepath.Join("../uploads", filepath.Base(filename))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(404).SendString("File not found")
		}
		return c.SendFile(filePath)

	})

	rolesDB := respository.NewRoleRepositoryDB(db)
	rolesService := service.NewRoleService(rolesDB)
	rolesHandler := handler.NewRolesHandler(rolesService)

	authsDB := respository.NewAuthRepositoryDB(db)
	authsService := service.NewAuthService(authsDB)
	authsHandler := handler.NewAuthHandler(authsService, db)

	usersDB := respository.NewUserRepositoryDB(db)
	usersService := service.NewUsersService(usersDB)
	usersHandler := handler.NewUsersHandler(usersService)

	productsDB := respository.NewProductRepositoryDB(db)
	productsService := service.NewProductService(productsDB)
	productsHandler := handler.NewProductHandler(productsService)
	// path.ProductPath(app, productsHandler)

	productCategoryDB := respository.NewProductCategoryRepositoryDB(db)
	productCategoryService := service.NewProductCategoryService(productCategoryDB)
	productCategoryHandler := handler.NewProductCategoryHandler(productCategoryService)

	rulesDB := respository.NewRulesRepositoryDB(db)
	rulesService := service.NewRulesService(rulesDB)
	rulesHandler := handler.NewRulesHandler(rulesService)

	installmentDB := respository.NewInstallmentRepositoryDB(db)
	installmentService := service.NewInstallmentService(installmentDB)
	installmentHandler := handler.NewInstallmentHandler(installmentService)

	fineDB := respository.NewFineRepositoryDB(db)
	fineService := service.NewFineService(fineDB)
	fineHandler := handler.NewFineHandler(fineService)

	fineCategoryDB := respository.NewFineCategoryRepositoryDB(db)
	fineCategoryService := service.NewFineCategoryService(fineCategoryDB)
	fineCategoryHandler := handler.NewFineCategoryHandler(fineCategoryService)

	memberDB := respository.NewMemberRepositoryDB(db)
	memberService := service.NewMemberService(memberDB)
	memberHandler := handler.NewMemberHandler(memberService)

	billDB := respository.NewBillRepositoryDB(db)
	billService := service.NewBillService(billDB, productsDB, fineDB, installmentDB)
	billHandler := handler.NewBillHandler(billService)

	path.ProductCategoryPath(app, productCategoryHandler, authsService, usersService)
	path.RulesPath(app, rulesHandler, authsService, usersService)
	path.InstallmentPath(app, installmentHandler, authsService, usersService)
	path.FinePath(app, fineHandler, authsService, usersService)
	path.FineCategoryPath(app, fineCategoryHandler, authsService, usersService)
	path.MemberPath(app, memberHandler, authsService, usersService)
	path.BillPath(app, billHandler, authsService, usersService)
	path.ProductPath(app, productsHandler, authsService, usersService)
	path.RolesPath(app, rolesHandler, authsService, usersService)
	path.UsersPath(app, usersHandler, authsService, usersService)
	path.AuthPath(app, authsHandler, authsService, usersService)

	backup.StartAutoBackupScheduler()

	// สมมติว่า s คือ billService
	// สร้าง "พนักงาน" เพียงคนเดียวที่ทำงานทุก 1 นาที
	// go func() {
	// 	log.Println("🚀 Starting Daily Update Cron Job...")
	// 	// กำหนดให้ทำงานทุก 1 นาที (หรือเปลี่ยนเป็น "0 1 * * *" เพื่อให้ทำงานตอนตี 1 ของทุกวัน)
	// 	ticker := time.NewTicker(1 * time.Minute)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		log.Println("==================================================")
	// 		log.Println("🔄 Cron Job Triggered: Starting daily bill updates...")
	// 		log.Println("1️⃣ Calling UpdateDailyInterest()...")
	// 		if err := billService.UpdateDailyInterest(); err != nil {
	// 			log.Printf("❌ Cron Job Error in UpdateDailyInterest: %v", err)
	// 			continue
	// 		}
	// 		log.Println("✅ UpdateDailyInterest() finished.")
	// 		log.Println("2️⃣ Calling AutoApplyInstallementLateFees()...")
	// 		if err := billService.AutoApplyInstallementLateFees(); err != nil {
	// 			log.Printf("❌ Cron Job Error in AutoApplyInstallementLateFees: %v", err)
	// 			continue
	// 		}
	// 		log.Println("✅ AutoApplyInstallementLateFees() finished.")
	// 		log.Println("🎉 Daily bill updates completed successfully.")
	// 		log.Println("==================================================")
	// 	}
	// }()
	go func() {
		log.Println("🚀 Starting Daily Update Cron Job at midnight (Asia/Bangkok)...")

		location, err := time.LoadLocation("Asia/Bangkok")
		if err != nil {
			log.Fatalf("❌ Failed to load timezone: %v", err)
		}

		for {
			now := time.Now().In(location)

			// คำนวณเวลาเที่ยงคืนของวันถัดไป
			nextMidnight := time.Date(
				now.Year(), now.Month(), now.Day()+1,
				0, 0, 0, 0, location,
			)

			durationUntilMidnight := time.Until(nextMidnight)
			log.Printf("⏳ Sleeping until midnight (in %v)...", durationUntilMidnight)

			time.Sleep(durationUntilMidnight)

			// 🔁 เริ่มรัน Job ตอนเที่ยงคืน
			log.Println("==================================================")
			log.Println("🔄 Cron Job Triggered: Starting daily bill updates...")

			log.Println("1️⃣ Calling UpdateDailyInterest()...")
			if err := billService.UpdateDailyInterest(); err != nil {
				log.Printf("❌ Cron Job Error in UpdateDailyInterest: %v", err)
				continue
			}
			log.Println("✅ UpdateDailyInterest() finished.")

			log.Println("2️⃣ Calling AutoApplyInstallementLateFees()...")
			if err := billService.AutoApplyInstallementLateFees(); err != nil {
				log.Printf("❌ Cron Job Error in AutoApplyInstallementLateFees: %v", err)
				continue
			}
			log.Println("✅ AutoApplyInstallementLateFees() finished.")

			log.Println("🎉 Daily bill updates completed successfully.")
			log.Println("==================================================")

			// วนลูปรอถึงเที่ยงคืนของวันถัดไป
		}
	}()

	app.Listen(":" + fmt.Sprint(viper.GetInt("PORT")))

}
