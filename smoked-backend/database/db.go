package database

import (
	"log"
	"smoked-meat/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func NewPostgresDB() error {
	dsn := "host=postgres port=5432 user=arthur password=hondadio dbname=meatdb sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	err = DB.AutoMigrate(&models.Assortment{}, models.User{}, &models.Order{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	var count int64
	DB.Model(&models.Assortment{}).Count(&count)
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
			log.Fatalf("Failed to insert initial data: %v", err)
		}
	}
	return nil
}

func GetDB() *gorm.DB {

	return DB
}
