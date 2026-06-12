package models

import "time"

// InventoryItem tracks how many of a crop a player has harvested.
type InventoryItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	CropID    uint      `gorm:"not null" json:"crop_id"`
	Count     int       `gorm:"not null;default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
