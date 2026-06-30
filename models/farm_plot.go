package models

import "time"

// FarmPlot 代表 6x5 农场网格中的一个单元格。
type FarmPlot struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	PlotIndex int        `gorm:"not null" json:"plot_index"` // 0..29
	Unlocked  bool       `gorm:"default:false" json:"unlocked"`
	LandLevel int        `gorm:"default:0" json:"land_level"` // 0=草地,1=黄土,2=黄土2,3=红土,4=黑土
	LandWork  int        `gorm:"default:0" json:"land_work"`
	CropID    *int       `json:"crop_id"`                   // nil = 空, 指向 CROPS 的从 0 开始的索引
	Progress  float64    `gorm:"default:0" json:"progress"` // 0.0 ~ 1.0
	WetTimer  float64    `gorm:"default:0" json:"wet_timer"`
	PlantedAt *time.Time `json:"planted_at"`

	// 照料系统字段
	WaterState        int     `gorm:"default:0" json:"water_state"`         // 0=正常,1=干燥,2=已浇水
	DryTimer          float64 `gorm:"default:0" json:"dry_timer"`           // 干燥秒数（产量惩罚）
	WaterProtectUntil float64 `gorm:"default:0" json:"water_protect_until"` // Unix 秒数
	BugCount          int     `gorm:"default:0" json:"bug_count"`
	BugSince          float64 `gorm:"default:0" json:"bug_since"`
	BugProtectUntil   float64 `gorm:"default:0" json:"bug_protect_until"`
	WeedCount         int     `gorm:"default:0" json:"weed_count"`
	WeedSince         float64 `gorm:"default:0" json:"weed_since"`
	WeedProtectUntil  float64 `gorm:"default:0" json:"weed_protect_until"`

	// 肥料状态
	FertUsed      int    `gorm:"default:0" json:"fert_used"`
	FertStageUsed string `gorm:"type:text;default:'{}'" json:"fert_stage_used"` // JSON 字典 例如 {"1":1}
	FertIDsUsed   string `gorm:"type:text;default:'[]'" json:"fert_ids_used"`   // JSON 数组 例如 [0,3]

	// 产量修饰符
	YieldBonusRate float64 `gorm:"default:0" json:"yield_bonus_rate"`
	YieldLossRate  float64 `gorm:"default:0" json:"yield_loss_rate"`

	LastProcessedAt *time.Time `json:"last_processed_at"` // 上次计算时间
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
