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
var CROPS = []CropSeed{
	{Key: "lettuce", NameZH: "生菜", NameEN: "Lettuce", SeedCost: 10, SellPrice: 20, GrowTime: 12, ExpReward: 5},
	{Key: "pepper", NameZH: "辣椒", NameEN: "Pepper", SeedCost: 20, SellPrice: 40, GrowTime: 30, ExpReward: 10},
	{Key: "eggplant", NameZH: "茄子", NameEN: "Eggplant", SeedCost: 30, SellPrice: 55, GrowTime: 45, ExpReward: 15},
	{Key: "tomato", NameZH: "西红柿", NameEN: "Tomato", SeedCost: 40, SellPrice: 75, GrowTime: 60, ExpReward: 20},
	{Key: "strawberry", NameZH: "草莓", NameEN: "Strawberry", SeedCost: 50, SellPrice: 100, GrowTime: 80, ExpReward: 25},
	{Key: "corn", NameZH: "玉米", NameEN: "Corn", SeedCost: 60, SellPrice: 120, GrowTime: 100, ExpReward: 30},
	{Key: "sunflower", NameZH: "向日葵", NameEN: "Sunflower", SeedCost: 80, SellPrice: 160, GrowTime: 140, ExpReward: 40},
	{Key: "pumpkin", NameZH: "南瓜", NameEN: "Pumpkin", SeedCost: 100, SellPrice: 200, GrowTime: 180, ExpReward: 50},
	{Key: "watermelon", NameZH: "西瓜", NameEN: "Watermelon", SeedCost: 130, SellPrice: 260, GrowTime: 230, ExpReward: 60},
}
