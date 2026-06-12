package dto

// SaveRequest is the full game save sent by the Godot client.
type SaveRequest struct {
	Gold           int            `json:"gold"`
	Level          int            `json:"level"`
	ExpVal         int            `json:"exp_val"`
	ExpToLevel     int            `json:"exp_to_level"`
	GameTime       float64        `json:"game_time"`
	SelectedSeed   int            `json:"selected_seed"`
	ToolMode       int            `json:"tool_mode"`
	SavedAt        int64          `json:"saved_at"`
	Plots          []PlotData     `json:"plots"`
	Inventory      map[string]int `json:"inventory"`
	FertilizerInv  map[string]int `json:"fertilizer_inventory"`
	SelectedFert   int            `json:"selected_fertilizer"`
}

// PlotData represents a single farm plot in the save payload.
type PlotData struct {
	PlotIndex        int     `json:"plot_index"`
	Unlocked         bool    `json:"unlocked"`
	LandLevel        int     `json:"land_level"`
	LandWork         int     `json:"land_work"`
	CropID           *int    `json:"crop_id"`
	Progress         float64 `json:"progress"`
	WetTimer         float64 `json:"wet_timer"`
	WaterState       int     `json:"water_state"`
	DryTimer         float64 `json:"dry_timer"`
	WaterProtectUntil float64 `json:"water_protect_until"`
	BugCount         int     `json:"bug_count"`
	BugSince         float64 `json:"bug_since"`
	BugProtectUntil  float64 `json:"bug_protect_until"`
	WeedCount        int     `json:"weed_count"`
	WeedSince        float64 `json:"weed_since"`
	WeedProtectUntil float64 `json:"weed_protect_until"`
	FertUsed         int     `json:"fert_used"`
	FertStageUsed    string  `json:"fert_stage_used"`
	FertIDsUsed      string  `json:"fert_ids_used"`
	YieldBonusRate   float64 `json:"yield_bonus_rate"`
	YieldLossRate    float64 `json:"yield_loss_rate"`
}

// LoadResponse is the full game save returned by the server.
// It has the same shape as SaveRequest so the client can parse it identically.
type LoadResponse = SaveRequest

// SellRequest is sent when the client sells inventory items.
type SellRequest struct {
	CropID int `json:"crop_id" binding:"required"`
	Count  int `json:"count" binding:"required,min=1"`
}

// SellResponse is returned after a successful sell.
type SellResponse struct {
	Gold       int `json:"gold"`
	SoldCount  int `json:"sold_count"`
	GoldEarned int `json:"gold_earned"`
}

// ActionRequest is a unified request for all farm actions.
type ActionRequest struct {
	Action  string `json:"action" binding:"required"` // water|fertilize|harvest|remove_bug|remove_weed|plant|shovel|harvest_all
	PlotIndex *int  `json:"plot_index,omitempty"`    // for single-tile actions
	CropID    *int  `json:"crop_id,omitempty"`       // for plant
	FertID    *int  `json:"fert_id,omitempty"`       // for fertilize
	Count     *int  `json:"count,omitempty"`         // for sell
}

// ActionResponse is returned after every successful action.
// Contains the full farm state so the client can re-render.
type ActionResponse struct {
	*LoadResponse
	Message string `json:"message,omitempty"` // e.g. "收获 生菜 x4!"
}
