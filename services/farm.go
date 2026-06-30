package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"gorm.io/gorm"
)

// SaveFarm 仅持久化非权威的客户端首选项。
// 所有游戏状态（金币、等级、地块、库存）都是服务器权威的，并且仅通过
// /farm/action 专门发生变更——客户端不能在这里覆盖它。
func SaveFarm(userID uint, req dto.SaveRequest) error {
	userUpdates := map[string]interface{}{
		"selected_seed":       req.SelectedSeed,
		"tool_mode":           req.ToolMode,
		"selected_fertilizer": req.SelectedFert,
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
		return fmt.Errorf("update user prefs: %w", err)
	}
	return nil
}

func LoadFarmConfig() *dto.FarmConfigResponse {
	crops := make([]dto.CropConfigDTO, 0, len(CROP_CONFIGS))
	for i, cfg := range CROP_CONFIGS {
		crops = append(crops, dto.CropConfigDTO{
			ID:         i,
			Name:       cfg.Name,
			TextureKey: cfg.Key,
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

// LoadFarm 检索用户的完整保存状态。
// 首先运行 ProcessFarm 来计算经过时间的变化。
// 它获取了每个用户的写入锁，因为 ProcessFarm 会改变状态。
func LoadFarm(userID uint) (*dto.LoadResponse, error) {
	defer lockUser(userID)()
	return loadFarmLocked(userID)
}

// loadFarmLocked 是 LoadFarm 的无锁实现主体。已经持有
// 每个用户锁的调用者（例如 ExecuteAction）必须使用它，以避免
// 在非重入的 sync.Mutex 上发生重入死锁。
func loadFarmLocked(userID uint) (*dto.LoadResponse, error) {
	if err := ProcessFarm(userID); err != nil {
		return nil, fmt.Errorf("process farm: %w", err)
	}
	// 1. 获取用户
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// 2. 获取所有农场块
	var plots []models.FarmPlot
	if err := database.DB.Where("user_id = ?", userID).Order("plot_index").Find(&plots).Error; err != nil {
		return nil, fmt.Errorf("load plots: %w", err)
	}

	plotData := make([]dto.PlotData, len(plots))
	now := time.Now()
	for i, p := range plots {
		plantedAt := unixSeconds(p.PlantedAt)
		plotData[i] = dto.PlotData{
			PlotIndex:         p.PlotIndex,
			Unlocked:          p.Unlocked,
			LandLevel:         p.LandLevel,
			LandWork:          p.LandWork,
			CropID:            p.CropID,
			PlantedAt:         plantedAt,
			EstimatedMatureAt: estimateMatureAt(p, now),
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

	// 3. 获取库存
	var items []models.InventoryItem
	if err := database.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("load inventory: %w", err)
	}
	inv := make(map[string]int)
	for _, item := range items {
		inv[strconv.Itoa(int(item.CropID))] = item.Count
	}

	// 4. 获取肥料库存
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
		ServerTime:    unixSeconds(&now),
		SelectedSeed:  user.SelectedSeed,
		ToolMode:      user.ToolMode,
		Plots:         plotData,
		Inventory:     inv,
		FertilizerInv: fertInv,
		SelectedFert:  user.SelectedFert,
	}, nil
}

// SellCrop 在服务器端出售库存物品（防作弊）。
func SellCrop(userID uint, cropID int, count int) (*dto.SellResponse, error) {
	defer lockUser(userID)()
	return sellCropLocked(userID, cropID, count)
}

func sellCropLocked(userID uint, cropID int, count int) (*dto.SellResponse, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}
	var item models.InventoryItem
	err := database.DB.Where("user_id = ? AND crop_id = ?", userID, cropID).First(&item).Error
	if err != nil {
		return nil, errors.New("no such crop in inventory")
	}
	if item.Count < count {
		return nil, errors.New("not enough items")
	}

	// 从 CROP_CONFIGS 查找售价（对应前端 CROPS[cid][6]）
	if cropID < 0 || cropID >= len(CROP_CONFIGS) {
		return nil, fmt.Errorf("invalid crop id: %d", cropID)
	}
	goldEarned := CROP_CONFIGS[cropID].UnitSell * count

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 扣除库存
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
		// 增加金币
		return tx.Model(&models.User{}).Where("id = ?", userID).
			Update("gold", gorm.Expr("gold + ?", goldEarned)).Error
	})
	if err != nil {
		return nil, err
	}

	// 返回更新后的金币
	var user models.User
	database.DB.First(&user, userID)

	return &dto.SellResponse{
		Gold:       user.Gold,
		SoldCount:  count,
		GoldEarned: goldEarned,
	}, nil
}
