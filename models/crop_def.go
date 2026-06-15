package models

// CropDef is the master definition for each crop type.
// Seed in bulk at startup from the CROPS constant below.
type CropDef struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Key        string  `gorm:"uniqueIndex;size:32;not null" json:"key"`       // e.g. "tomato"
	NameZH     string  `gorm:"size:32;not null" json:"name_zh"`              // e.g. "西红柿"
	NameEN     string  `gorm:"size:32;not null" json:"name_en"`              // e.g. "Tomato"
	SeedCost   int     `gorm:"not null" json:"seed_cost"`
	SellPrice  int     `gorm:"not null" json:"sell_price"`
	GrowTime   float64 `gorm:"not null" json:"grow_time"` // seconds
	ExpReward  int     `gorm:"not null" json:"exp_reward"`
}

// CropSeed mirrors CropDef for seeding the database.
type CropSeed struct {
	Key       string
	NameZH    string
	NameEN    string
	SeedCost  int
	SellPrice int
	GrowTime  float64
	ExpReward int
}

// CROPS is the master crop table — must stay in sync with the Godot client.
// SellPrice = per-unit sell price (matches frontend CROPS[cid][6] / CROP_CONFIGS.UnitSell).
var CROPS = []CropSeed{
	{Key: "lettuce", NameZH: "生菜", NameEN: "Lettuce", SeedCost: 12, SellPrice: 8, GrowTime: 12, ExpReward: 5},
	{Key: "pepper", NameZH: "辣椒", NameEN: "Pepper", SeedCost: 20, SellPrice: 10, GrowTime: 20, ExpReward: 10},
	{Key: "eggplant", NameZH: "茄子", NameEN: "Eggplant", SeedCost: 35, SellPrice: 19, GrowTime: 32, ExpReward: 15},
	{Key: "tomato", NameZH: "西红柿", NameEN: "Tomato", SeedCost: 55, SellPrice: 19, GrowTime: 48, ExpReward: 20},
	{Key: "strawberry", NameZH: "草莓", NameEN: "Strawberry", SeedCost: 80, SellPrice: 18, GrowTime: 70, ExpReward: 25},
	{Key: "corn", NameZH: "玉米", NameEN: "Corn", SeedCost: 120, SellPrice: 57, GrowTime: 100, ExpReward: 30},
	{Key: "sunflower", NameZH: "向日葵", NameEN: "Sunflower", SeedCost: 170, SellPrice: 125, GrowTime: 135, ExpReward: 40},
	{Key: "pumpkin", NameZH: "南瓜", NameEN: "Pumpkin", SeedCost: 240, SellPrice: 240, GrowTime: 180, ExpReward: 50},
	{Key: "watermelon", NameZH: "西瓜", NameEN: "Watermelon", SeedCost: 320, SellPrice: 196, GrowTime: 230, ExpReward: 60},
}
