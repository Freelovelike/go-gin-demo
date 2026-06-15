package models

import "time"

// FertilizerInventory tracks how many of each fertilizer a player owns.
type FertilizerInventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	FertIndex int       `gorm:"not null" json:"fert_index"` // FERTILIZERS array index (0-6)
	Count     int       `gorm:"not null;default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
