package models

import "time"

// FertilizerInventory 记录玩家拥有每种肥料的数量。
type FertilizerInventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	FertIndex int       `gorm:"not null" json:"fert_index"` // FERTILIZERS 数组索引 (0-6)
	Count     int       `gorm:"not null;default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
