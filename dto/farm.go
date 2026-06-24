package dto

// SaveRequest carries the only client-owned, non-authoritative preferences the
// server persists via /farm/save. All real game state (gold, level, plots,
// inventory) is server-authoritative and is mutated exclusively through
// /farm/action — the client cannot write it here. The client may still POST a
// larger body; extra fields are simply ignored during JSON binding.
type SaveRequest struct {
	SelectedSeed int `json:"selected_seed"`
	ToolMode     int `json:"tool_mode"`
	SelectedFert int `json:"selected_fertilizer"`
}

// PlotData represents a single farm plot in the save payload.
type PlotData struct {
	PlotIndex         int            `json:"plot_index"`
	Unlocked          bool           `json:"unlocked"`
	LandLevel         int            `json:"land_level"`
	LandWork          int            `json:"land_work"`
	CropID            *int           `json:"crop_id"`
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

// LoadResponse is the full game save returned by the server on load and after
// every action. It is server-authoritative — the client renders it verbatim.
type LoadResponse struct {
	Gold          int            `json:"gold"`
	Level         int            `json:"level"`
	ExpVal        int            `json:"exp_val"`
	ExpToLevel    int            `json:"exp_to_level"`
	GameTime      float64        `json:"game_time"`
	SelectedSeed  int            `json:"selected_seed"`
	ToolMode      int            `json:"tool_mode"`
	Plots         []PlotData     `json:"plots"`
	Inventory     map[string]int `json:"inventory"`
	FertilizerInv map[string]int `json:"fertilizer_inventory"`
	SelectedFert  int            `json:"selected_fertilizer"`
}

// SellRequest is sent when the client sells inventory items.
type SellRequest struct {
	CropID *int `json:"crop_id" binding:"required"`
	Count  int  `json:"count" binding:"required,min=1"`
}

// SellResponse is returned after a successful sell.
type SellResponse struct {
	Gold       int `json:"gold"`
	SoldCount  int `json:"sold_count"`
	GoldEarned int `json:"gold_earned"`
}

// ActionRequest is a unified request for all farm actions.
type ActionRequest struct {
	Action    string `json:"action" binding:"required"` // water|fertilize|harvest|remove_bug|remove_weed|plant|shovel|harvest_all
	PlotIndex *int   `json:"plot_index,omitempty"`      // for single-tile actions
	CropID    *int   `json:"crop_id,omitempty"`         // for plant
	FertID    *int   `json:"fert_id,omitempty"`         // for fertilize
	Count     *int   `json:"count,omitempty"`           // for sell
}

// ActionResponse is returned after every successful action.
// Contains the full farm state so the client can re-render.
type ActionResponse struct {
	*LoadResponse
	Message string `json:"message,omitempty"` // e.g. "收获 生菜 x4!"
}

// FarmConfigResponse exposes server-authoritative crop/fertilizer definitions
// so the client does not hardcode gameplay configuration locally.
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
