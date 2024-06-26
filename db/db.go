package db

import (
	"fmt"
	"log"
	"os"

	"github.com/ruined/yamb/v1/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=UTC",
		os.Getenv("YAMB_PQHOST"),
		os.Getenv("YAMB_PQUSERNAME"),
		os.Getenv("YAMB_PQPASSWORD"),
		"yamb",
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(&models.Artist{}, &models.LatestRelease{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}
