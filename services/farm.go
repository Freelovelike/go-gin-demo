package services

import (
	"errors"
	"fmt"
	"strconv"

	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"gorm.io/gorm"
)

// SaveFarm performs a full save: updates User + replaces all 30 FarmPlots + replaces Inventory.
func SaveFarm(userID uint, req dto.SaveRequest) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Update User fields
		userUpdates := map[string]interface{}{
			"gold":          req.Gold,
			"level":         req.Level,
			"exp_val":       req.ExpVal,
			"exp_to_lvl":    req.ExpToLevel,
			"game_time":     req.GameTime,
			"selected_seed": req.SelectedSeed,
			"tool_mode":     req.ToolMode,
		}
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		// 2. Upsert each FarmPlot
		for _, p := range req.Plots {
			plot := models.FarmPlot{
				UserID:            userID,
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
				FertStageUsed:     p.FertStageUsed,
				FertIDsUsed:       p.FertIDsUsed,
				YieldBonusRate:    p.YieldBonusRate,
				YieldLossRate:     p.YieldLossRate,
			}
			result := tx.Where("user_id = ? AND plot_index = ?", userID, p.PlotIndex).
				Assign(plot).
				FirstOrCreate(&models.FarmPlot{})
			if result.Error != nil {
				return fmt.Errorf("upsert plot %d: %w", p.PlotIndex, result.Error)
			}
		}

		// 3. Replace Inventory: delete old, insert new
		if err := tx.Where("user_id = ?", userID).Delete(&models.InventoryItem{}).Error; err != nil {
			return fmt.Errorf("delete old inventory: %w", err)
		}
		for cropIDStr, count := range req.Inventory {
			if count <= 0 {
				continue
			}
			cropID, err := strconv.Atoi(cropIDStr)
			if err != nil {
				continue
			}
			item := models.InventoryItem{
				UserID: userID,
				CropID: uint(cropID),
				Count:  count,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("insert inventory crop %d: %w", cropID, err)
			}
		}

		return nil
	})
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
			FertStageUsed:     p.FertStageUsed,
			FertIDsUsed:       p.FertIDsUsed,
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

	return &dto.LoadResponse{
		Gold:           user.Gold,
		Level:          user.Level,
		ExpVal:         user.ExpVal,
		ExpToLevel:     user.ExpToLvl,
		GameTime:       user.GameTime,
		SelectedSeed:   user.SelectedSeed,
		ToolMode:       user.ToolMode,
		Plots:          plotData,
		Inventory:      inv,
		FertilizerInv:  map[string]int{}, // TODO: add fertilizer_inventory table
		SelectedFert:   -1,
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

	// Look up sell price from CropDef
	var cropDef models.CropDef
	if err := database.DB.Where("id = ?", cropID).First(&cropDef).Error; err != nil {
		return nil, errors.New("crop definition not found")
	}

	goldEarned := cropDef.SellPrice * count

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
