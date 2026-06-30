package dto

// SaveRequest 携带服务器通过 /farm/save 持久化的唯一属于客户端的、非权威的首选项。
// 所有真正的游戏状态（金币、等级、地块、库存）都是服务器权威的，只能专门通过
// /farm/action 进行变更——客户端不能在这里写入它。客户端仍然可以 POST 一个
// 更大的主体；多余的字段在 JSON 绑定期间将被直接忽略。
type SaveRequest struct {
	SelectedSeed int `json:"selected_seed"`
	ToolMode     int `json:"tool_mode"`
	SelectedFert int `json:"selected_fertilizer"`
}

// PlotData 代表存档有效载荷中的单个农场块。
type PlotData struct {
	PlotIndex         int            `json:"plot_index"`
	Unlocked          bool           `json:"unlocked"`
	LandLevel         int            `json:"land_level"`
	LandWork          int            `json:"land_work"`
	CropID            *int           `json:"crop_id"`
	PlantedAt         float64        `json:"planted_at"`
	EstimatedMatureAt float64        `json:"estimated_mature_at"`
	Progress          float64        `json:"progress"`
	WetTimer          float64        `json:"wet_timer"`
	WaterState        int            `json:"water_state"`
	DryTimer          float64        `json:"dry_timer"`
	WaterProtectUntil float64        `json:"water_protect_until"`
	BugCount          int            `json:"bug_count"`
	BugSince          float64        `json:"bug_since"`
	BugProtectUntil   float64        `json:"bug_protect_until"`
	WeedCount         int            `json:"weed_count"`
	WeedSince         float64        `json:"weed_since"`
	WeedProtectUntil  float64        `json:"weed_protect_until"`
	FertUsed          int            `json:"fert_used"`
	FertStageUsed     map[string]int `json:"fert_stage_used"`
	FertIDsUsed       []int          `json:"fert_ids_used"`
	YieldBonusRate    float64        `json:"yield_bonus_rate"`
	YieldLossRate     float64        `json:"yield_loss_rate"`
}

// LoadResponse 是服务器在加载时和每次操作后返回的完整游戏存档。
// 它是服务器权威的——客户端逐字（如实）渲染它。
type LoadResponse struct {
	Gold          int            `json:"gold"`
	Level         int            `json:"level"`
	ExpVal        int            `json:"exp_val"`
	ExpToLevel    int            `json:"exp_to_level"`
	GameTime      float64        `json:"game_time"`
	ServerTime    float64        `json:"server_time"`
	SelectedSeed  int            `json:"selected_seed"`
	ToolMode      int            `json:"tool_mode"`
	Plots         []PlotData     `json:"plots"`
	Inventory     map[string]int `json:"inventory"`
	FertilizerInv map[string]int `json:"fertilizer_inventory"`
	SelectedFert  int            `json:"selected_fertilizer"`
}

// SellRequest 在客户端出售库存物品时发送。
type SellRequest struct {
	CropID *int `json:"crop_id" binding:"required"`
	Count  int  `json:"count" binding:"required,min=1"`
}

// SellResponse 在成功出售后返回。
type SellResponse struct {
	Gold       int `json:"gold"`
	SoldCount  int `json:"sold_count"`
	GoldEarned int `json:"gold_earned"`
}

// ActionRequest 是对所有农场操作的统一请求。
type ActionRequest struct {
	Action    string `json:"action" binding:"required"` // water|fertilize|harvest|remove_bug|remove_weed|plant|shovel|harvest_all
	PlotIndex *int   `json:"plot_index,omitempty"`      // 用于单瓦片操作
	CropID    *int   `json:"crop_id,omitempty"`         // 用于种植
	FertID    *int   `json:"fert_id,omitempty"`         // 用于施肥
	Count     *int   `json:"count,omitempty"`           // 用于出售
}

// ActionResponse 在每次成功操作后返回。
// 包含完整的农场状态以便客户端重新渲染。
type ActionResponse struct {
	*LoadResponse
	Message string `json:"message,omitempty"` // 例如 "收获 生菜 x4!"
}

// FarmConfigResponse 公开服务器权威的作物/肥料定义，
// 以便客户端不必在本地硬编码游戏玩法配置。
type FarmConfigResponse struct {
	Crops                 []CropConfigDTO       `json:"crops"`
	Fertilizers           []FertilizerConfigDTO `json:"fertilizers"`
	StageThresholds       StageThresholdsDTO    `json:"stage_thresholds"`
	RenderStageThresholds []float64             `json:"render_stage_thresholds"`
}

type CropConfigDTO struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	TextureKey string  `json:"texture_key"`
	SeedCost   int     `json:"seed_cost"`
	GrowTime   float64 `json:"grow_time"`
	BaseYield  int     `json:"base_yield"`
	UnitSell   int     `json:"unit_sell"`
	MinYield   int     `json:"min_yield"`
	MaxYield   int     `json:"max_yield"`
	DryRate    float64 `json:"dry_rate"`
	BugRate    float64 `json:"bug_rate"`
	WeedRate   float64 `json:"weed_rate"`
	MaxBug     int     `json:"max_bug"`
	MaxWeed    int     `json:"max_weed"`
}

type FertilizerConfigDTO struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Cost            int     `json:"cost"`
	Type            string  `json:"type"`
	EffectValue     float64 `json:"effect_value"`
	AllowedStages   []int   `json:"allowed_stages"`
	PerCropLimit    int     `json:"per_crop_limit"`
	MaxMinutesLimit int     `json:"max_minutes_limit"`
}

type StageThresholdsDTO struct {
	SeedEnd    float64 `json:"seed_end"`
	SproutEnd  float64 `json:"sprout_end"`
	GrowingEnd float64 `json:"growing_end"`
}
