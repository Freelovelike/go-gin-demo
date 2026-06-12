package models

import "time"

// FarmPlot represents one cell in the 6x5 farm grid.
type FarmPlot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	PlotIndex int       `gorm:"not null" json:"plot_index"` // 0..29
	Unlocked  bool      `gorm:"default:false" json:"unlocked"`
	LandLevel int       `gorm:"default:0" json:"land_level"` // 0=grass,1=yellow,2=yellow2,3=red,4=black
	LandWork  int       `gorm:"default:0" json:"land_work"`
	CropID    *int      `json:"crop_id"` // nil = empty, 0-based index into CROPS
	Progress  float64   `gorm:"default:0" json:"progress"` // 0.0 ~ 1.0
	WetTimer  float64   `gorm:"default:0" json:"wet_timer"`
	PlantedAt *time.Time `json:"planted_at"`

	// Care system fields
	WaterState        int     `gorm:"default:0" json:"water_state"`         // 0=Normal,1=Dry,2=Watered
	DryTimer          float64 `gorm:"default:0" json:"dry_timer"`           // seconds spent dry (yield penalty)
	WaterProtectUntil float64 `gorm:"default:0" json:"water_protect_until"` // game-time seconds
	BugCount          int     `gorm:"default:0" json:"bug_count"`
	BugSince          float64 `gorm:"default:0" json:"bug_since"`
	BugProtectUntil   float64 `gorm:"default:0" json:"bug_protect_until"`
	WeedCount         int     `gorm:"default:0" json:"weed_count"`
	WeedSince         float64 `gorm:"default:0" json:"weed_since"`
	WeedProtectUntil  float64 `gorm:"default:0" json:"weed_protect_until"`

	// Fertilizer state
	FertUsed    int    `gorm:"default:0" json:"fert_used"`
	FertStageUsed string `gorm:"type:text;default:'{}'" json:"fert_stage_used"` // JSON dict e.g. {"1":1}
	FertIDsUsed   string `gorm:"type:text;default:'[]'" json:"fert_ids_used"`   // JSON array e.g. [0,3]

	// Yield modifiers
	YieldBonusRate float64 `gorm:"default:0" json:"yield_bonus_rate"`
	YieldLossRate  float64 `gorm:"default:0" json:"yield_loss_rate"`

	LastProcessedAt *time.Time `json:"last_processed_at"` // 上次计算时间
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
