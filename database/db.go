package database

import (
	"fmt"
	"log"
	"neobank-lite/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to DB: ", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(&models.User{}, &models.Account{}, &models.Transaction{})
	if err != nil {
		log.Fatal("❌ Auto migration failed: ", err)
	}

	DB = db
	fmt.Println("✅ Connected to PostgreSQL successfully!")
}
