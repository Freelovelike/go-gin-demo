package models

import "time"

// User represents a registered player.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Password     string    `gorm:"size:128;not null" json:"-"`
	Gold         int       `gorm:"default:200" json:"gold"`
	Level        int       `gorm:"default:1" json:"level"`
	ExpVal       int       `gorm:"column:exp_val;default:0" json:"exp_val"`
	ExpToLvl     int       `gorm:"column:exp_to_lvl;default:100" json:"exp_to_lvl"`
	GameTime     float64   `gorm:"column:game_time;default:0" json:"game_time"`
	SelectedSeed int       `gorm:"column:selected_seed;default:-1" json:"selected_seed"`
	ToolMode     int       `gorm:"column:tool_mode;default:0" json:"tool_mode"`
	SelectedFert int       `gorm:"column:selected_fertilizer;default:-1" json:"selected_fertilizer"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
