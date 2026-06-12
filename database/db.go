package database

import (
	"log"

	"go-gin-demo/config"
	"go-gin-demo/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitPostgres(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("PostgreSQL connected")
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.FarmPlot{},
		&models.InventoryItem{},
		&models.CropDef{},
	)
	if err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}
	log.Println("Database migration completed")
}

func SeedCropDefs() {
	for _, cs := range models.CROPS {
		result := DB.Where("key = ?", cs.Key).FirstOrCreate(&models.CropDef{
			Key:       cs.Key,
			NameZH:    cs.NameZH,
			NameEN:    cs.NameEN,
			SeedCost:  cs.SeedCost,
			SellPrice: cs.SellPrice,
			GrowTime:  cs.GrowTime,
			ExpReward: cs.ExpReward,
		})
		if result.Error != nil {
			log.Printf("seed crop %s failed: %v", cs.Key, result.Error)
		}
	}
	log.Println("Crop definitions seeded")
}
