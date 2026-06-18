package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"gorm.io/gorm"
)

// SaveFarm persists only non-authoritative client preferences (selected seed, tool mode).
// All game state (gold, level, plots, inventory) is server-authoritative and is mutated
// exclusively through /farm/action — the client cannot overwrite it here.
func SaveFarm(userID uint, req dto.SaveRequest) error {
	userUpdates := map[string]interface{}{
		"selected_seed": req.SelectedSeed,
		"tool_mode":     req.ToolMode,
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
		return fmt.Errorf("update user prefs: %w", err)
	}
	return nil
}

func LoadFarmConfig() *dto.FarmConfigResponse {
	crops := make([]dto.CropConfigDTO, 0, len(CROP_CONFIGS))
	cropNames := []string{"生菜", "辣椒", "茄子", "西红柿", "草莓", "玉米", "向日葵", "南瓜", "西瓜"}
	cropKeys := []string{"lettuce", "pepper", "eggplant", "tomato", "strawberry", "corn", "sunflower", "pumpkin", "watermelon"}
	for i, cfg := range CROP_CONFIGS {
		crops = append(crops, dto.CropConfigDTO{
			ID:         i,
			Name:       cropNames[i],
			TextureKey: cropKeys[i],
			SeedCost:   cfg.SeedCost,
			GrowTime:   cfg.GrowTime,
			BaseYield:  cfg.BaseYield,
			UnitSell:   cfg.UnitSell,
			MinYield:   cfg.MinYield,
			MaxYield:   cfg.MaxYield,
			DryRate:    cfg.DryRate,
			BugRate:    cfg.BugRate,
			WeedRate:   cfg.WeedRate,
			MaxBug:     cfg.MaxBug,
			MaxWeed:    cfg.MaxWeed,
		})
	}

	fertilizers := make([]dto.FertilizerConfigDTO, 0, len(FERTILIZER_CONFIGS))
	for i, cfg := range FERTILIZER_CONFIGS {
		allowedStages := append([]int(nil), cfg.AllowedStages...)
		fertilizers = append(fertilizers, dto.FertilizerConfigDTO{
			ID:              i,
			Name:            cfg.Name,
			Cost:            cfg.Cost,
			Type:            cfg.Type,
			EffectValue:     cfg.EffectValue,
			AllowedStages:   allowedStages,
			PerCropLimit:    cfg.PerCropLimit,
			MaxMinutesLimit: cfg.MaxMinutesLimit,
		})
	}

	return &dto.FarmConfigResponse{
		Crops:       crops,
		Fertilizers: fertilizers,
		StageThresholds: dto.StageThresholdsDTO{
			SeedEnd:    stageSeedEnd,
			SproutEnd:  stageSproutEnd,
			GrowingEnd: stageGrowingEnd,
		},
		RenderStageThresholds: []float64{stageSeedEnd, stageSproutEnd, renderStageMidEnd, stageGrowingEnd},
	}
}

func parseFertStageUsed(raw string) map[string]int {
	result := map[string]int{}
	if raw == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]int{}
	}
	return result
}

func parseFertIDsUsed(raw string) []int {
	result := []int{}
	if raw == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []int{}
	}
	return result
}

func encodeFertStageUsed(value map[string]int) string {
	if value == nil {
		return "{}"
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func encodeFertIDsUsed(value []int) string {
	if value == nil {
		return "[]"
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

// LoadFarm retrieves the full save state for a user.
// First runs ProcessFarm to calculate elapsed time changes.
func LoadFarm(userID uint) (*dto.LoadResponse, error) {
	if err := ProcessFarm(userID); err != nil {
		return nil, fmt.Errorf("process farm: %w", err)
	}
	// 1. Get User
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// 2. Get all FarmPlots
	var plots []models.FarmPlot
	if err := database.DB.Where("user_id = ?", userID).Order("plot_index").Find(&plots).Error; err != nil {
		return nil, fmt.Errorf("load plots: %w", err)
	}

	plotData := make([]dto.PlotData, len(plots))
	for i, p := range plots {
		plotData[i] = dto.PlotData{
			PlotIndex:         p.PlotIndex,
			Unlocked:          p.Unlocked,
			LandLevel:         p.LandLevel,
			LandWork:          p.LandWork,
			CropID:            p.CropID,
			Progress:          p.Progress,
			WetTimer:          p.WetTimer,
			WaterState:        p.WaterState,
			DryTimer:          p.DryTimer,
			WaterProtectUntil: p.WaterProtectUntil,
			BugCount:          p.BugCount,
			BugSince:          p.BugSince,
			BugProtectUntil:   p.BugProtectUntil,
			WeedCount:         p.WeedCount,
			WeedSince:         p.WeedSince,
			WeedProtectUntil:  p.WeedProtectUntil,
			FertUsed:          p.FertUsed,
			FertStageUsed:     parseFertStageUsed(p.FertStageUsed),
			FertIDsUsed:       parseFertIDsUsed(p.FertIDsUsed),
			YieldBonusRate:    p.YieldBonusRate,
			YieldLossRate:     p.YieldLossRate,
		}
	}

	// 3. Get Inventory
	var items []models.InventoryItem
	if err := database.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("load inventory: %w", err)
	}
	inv := make(map[string]int)
	for _, item := range items {
		inv[strconv.Itoa(int(item.CropID))] = item.Count
	}

	// 4. Get Fertilizer Inventory
	var fertItems []models.FertilizerInventory
	if err := database.DB.Where("user_id = ?", userID).Find(&fertItems).Error; err != nil {
		return nil, fmt.Errorf("load fertilizer inventory: %w", err)
	}
	fertInv := make(map[string]int)
	for _, fi := range fertItems {
		if fi.Count > 0 {
			fertInv[strconv.Itoa(fi.FertIndex)] = fi.Count
		}
	}

	return &dto.LoadResponse{
		Gold:          user.Gold,
		Level:         user.Level,
		ExpVal:        user.ExpVal,
		ExpToLevel:    user.ExpToLvl,
		GameTime:      user.GameTime,
		SelectedSeed:  user.SelectedSeed,
		ToolMode:      user.ToolMode,
		Plots:         plotData,
		Inventory:     inv,
		FertilizerInv: fertInv,
		SelectedFert:  -1,
	}, nil
}

// SellCrop sells inventory items server-side (anti-cheat).
func SellCrop(userID uint, cropID int, count int) (*dto.SellResponse, error) {
	var item models.InventoryItem
	err := database.DB.Where("user_id = ? AND crop_id = ?", userID, cropID).First(&item).Error
	if err != nil {
		return nil, errors.New("no such crop in inventory")
	}
	if item.Count < count {
		return nil, errors.New("not enough items")
	}

	// Look up sell price from CROP_CONFIGS (matches frontend CROPS[cid][6])
	if cropID < 0 || cropID >= len(CROP_CONFIGS) {
		return nil, fmt.Errorf("invalid crop id: %d", cropID)
	}
	goldEarned := CROP_CONFIGS[cropID].UnitSell * count

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Deduct inventory
		item.Count -= count
		if item.Count == 0 {
			if err := tx.Delete(&item).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}
		// Add gold
		return tx.Model(&models.User{}).Where("id = ?", userID).
			Update("gold", gorm.Expr("gold + ?", goldEarned)).Error
	})
	if err != nil {
		return nil, err
	}

	// Return updated gold
	var user models.User
	database.DB.First(&user, userID)

	return &dto.SellResponse{
		Gold:       user.Gold,
		SoldCount:  count,
		GoldEarned: goldEarned,
	}, nil
}
