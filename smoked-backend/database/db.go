package database

import (
	"fmt"
	"os"
	"smoked-meat/models"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func NewPostgresDB() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error

	// Ретраи для подключения (5 попыток)
	for i := range 5 {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Printf("[error] Failed to open GORM connection (attempt %d/5): %v", i+1, err)
			if i == 4 {
				return fmt.Errorf("failed to connect to PostgreSQL after 5 attempts: %w", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("[error] Failed to get sql.DB (attempt %d/5): %v", i+1, err)
			if i == 4 {
				return fmt.Errorf("failed to get sql.DB after 5 attempts: %w", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		err = sqlDB.Ping()
		if err != nil {
			log.Printf("[error] Failed to ping DB (attempt %d/5): %v", i+1, err)
			if i == 4 {
				return fmt.Errorf("failed to ping PostgreSQL after 5 attempts: %w", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		log.Info().Msg("[info] Successfully connected to PostgreSQL")
		break
	}

	err = DB.AutoMigrate(&models.Assortment{}, &models.User{}, &models.Order{})
	if err != nil {
		log.Printf("[error] Failed to migrate database: %v", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info().Msg("[info] Database migration completed")

	var count int64
	if err := DB.Model(&models.Assortment{}).Count(&count).Error; err != nil {
		log.Printf("[error] Failed to count assortments: %v", err)
		return fmt.Errorf("failed to count assortments: %w", err)
	}

	if count == 0 {
		assortment := []models.Assortment{
			{Meat: "Pork",
				Availability: true,
				Price:        1800,
				Spice:        models.Spice{Recipe1: "Honey, black pepper, soy sauce!", Recipe2: "Quince jam, red pepper, teriyaki sauce!"}},
			{Meat: "Beef",
				Availability: true,
				Price:        2800,
				Spice:        models.Spice{Recipe1: "Pear cider, aromatic dried herbs!", Recipe2: "Apple cider, lecho sauce!"}},
			{Meat: "Chicken",
				Availability: true,
				Price:        1500,
				Spice:        models.Spice{Recipe1: "Coke with lemongrass!", Recipe2: "Orange juice, paprika!"}},
			{Meat: "Turkey",
				Availability: true,
				Price:        2000,
				Spice:        models.Spice{Recipe1: "Basil, cherry caramel!", Recipe2: "Rosemary, marjoram, pomegranate caramel!"}},
			{Meat: "Mutton",
				Availability: true,
				Price:        2400,
				Spice:        models.Spice{Recipe1: "Mustard marinade, star anise, tomatoes!", Recipe2: "Kefir marinade, hot pepper mix, fresh mint!"}},
			{Meat: "Venison",
				Availability: true,
				Price:        5600,
				Spice:        models.Spice{Recipe1: "Mead with garlic and wild blackberry!", Recipe2: "Mead with ginger and wild currant!"}},
		}
		if err := DB.Create(&assortment).Error; err != nil {
			log.Printf("[error] Failed to insert initial data: %v", err)
			return fmt.Errorf("failed to insert initial data: %w", err)
		}
		log.Info().Msg("[info] Initial assortment data inserted")
	}

	var count2 int64
	if err := DB.Model(&models.User{}).Count(&count2).Error; err != nil {
		log.Printf("[error] Failed to count assortments: %v", err)
		return fmt.Errorf("failed to count assortments: %w", err)
	}

	if count2 == 0 {
		users := []models.User{
			{Username: "Arthur",
				Password:    HashPassword("tutu22"),
				PhoneNumber: "+79786666666",
				Role:        "owner",
				Name:        "Arthur",
			},
			{Username: "HondaDio",
				Password:    HashPassword("hahaha"),
				PhoneNumber: "+79787777777",
				Role:        "owner",
				Name:        "Honda",
			}}
		if err := DB.Create(&users).Error; err != nil {
			log.Printf("[error] Failed to insert initial data: %v", err)
			return fmt.Errorf("failed to insert initial data: %w", err)
		}
		log.Info().Msg("[info] Initial users data inserted")
	}

	return nil
}

func GetDB() *gorm.DB {

	return DB
}

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return ""
	}
	return string(hashedPassword)
}
