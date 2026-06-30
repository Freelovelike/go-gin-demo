package database

import (
	"log"

	"go-gin-demo/config"
	"go-gin-demo/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局的 GORM 数据库实例。
var DB *gorm.DB

// InitPostgres 初始化并连接 PostgreSQL 数据库。
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

// AutoMigrate 自动迁移数据库模式以匹配定义的模型。
func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.FarmPlot{},
		&models.InventoryItem{},
		&models.FertilizerInventory{},
		&models.CropDef{},
	)
	if err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}
	log.Println("Database migration completed")
}

// SeedCropDefs 使用预定义的作物常量向数据库中填充初始数据。
func SeedCropDefs() {
	for _, cs := range models.CROPS {
		result := DB.Where("key = ?", cs.Key).Assign(models.CropDef{
			SeedCost:  cs.SeedCost,
			SellPrice: cs.SellPrice,
			GrowTime:  cs.GrowTime,
			ExpReward: cs.ExpReward,
		}).FirstOrCreate(&models.CropDef{
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
