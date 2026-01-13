package database

import (
	"fmt"
	"log"
	"os"

	"disney/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// loadEnv loads .env file only in local development
func loadEnv() {
	if os.Getenv("RENDER") == "" {
		// local development only
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system env")
		}
	}
}

var DB *gorm.DB

func InitDB() {
	// Load .env file only in local development
	loadEnv()

	// Get SSL mode (default to disable for local, require for production)
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	// Build DSN from env variables
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		sslMode,
	)

	// db connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	DB = db
	log.Println("Database connected")

	// Auto migrate tables
	DB.AutoMigrate(
		&models.User{},
		&models.Genre{},
		&models.AgeGroup{},
		&models.Cartoon{},
		&models.Character{},
		&models.Rating{},
		&models.Favourite{},
		&models.View{},
		&models.AdminLog{},
		&models.RequestLog{},
		&models.TimeTable{},
	)

	log.Println("Migration done")

	// Fix sequence issues after migration
	fixSequences()

	// Seed default genres and age groups if they don't exist
	seedDefaultData()
}

// fixSequences resets auto-increment sequences to avoid primary key conflicts
func fixSequences() {
	log.Println("Fixing auto-increment sequences...")

	// List of tables with their sequence names
	tables := map[string]string{
		"users":        "users_id_seq",
		"genres":       "genres_id_seq",
		"age_groups":   "age_groups_id_seq",
		"cartoons":     "cartoons_id_seq",
		"characters":   "characters_id_seq",
		"ratings":      "ratings_id_seq",
		"favourites":   "favourites_id_seq",
		"views":        "views_id_seq",
		"admin_logs":   "admin_logs_id_seq",
		"request_logs": "request_logs_id_seq",
		"time_tables":  "time_tables_id_seq",
	}

	for table, sequence := range tables {
		// Reset sequence to current max ID + 1
		query := fmt.Sprintf(`
			SELECT setval('%s', COALESCE((SELECT MAX(id) FROM %s), 0) + 1, false);
		`, sequence, table)

		if err := DB.Exec(query).Error; err != nil {
			log.Printf("Warning: Could not fix sequence for %s: %v", table, err)
		} else {
			log.Printf("Fixed sequence for %s", table)
		}
	}

	log.Println("Sequence fixing completed")
}

// seedDefaultData creates default genres and age groups if they don't exist
func seedDefaultData() {
	// Check if genres exist
	var genreCount int64
	DB.Model(&models.Genre{}).Count(&genreCount)

	if genreCount == 0 {
		log.Println("Seeding default genres...")
		genres := []models.Genre{
			{Name: "Action"},
			{Name: "Comedy"},
			{Name: "Drama"},
			{Name: "Romance"},
			{Name: "Thriller"},
			{Name: "Horror"},
			{Name: "Sci-Fi"},
			{Name: "Fantasy"},
		}

		for _, genre := range genres {
			DB.Create(&genre)
		}
		log.Printf("Created %d default genres", len(genres))
	}

	// Check if age groups exist
	var ageGroupCount int64
	DB.Model(&models.AgeGroup{}).Count(&ageGroupCount)

	if ageGroupCount == 0 {
		log.Println("Seeding default age groups...")
		ageGroups := []models.AgeGroup{
			{Label: "Preschool (2-4 years)"},
			{Label: "Kids (5-8 years)"},
			{Label: "Tweens (9-12 years)"},
			{Label: "Teens (13-17 years)"},
			{Label: "Adults (18+ years)"},
		}

		for _, ageGroup := range ageGroups {
			DB.Create(&ageGroup)
		}
		log.Printf("Created %d default age groups", len(ageGroups))
	}
}
