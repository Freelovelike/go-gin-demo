package services

import (
	"fmt"
	"math/rand"
	"time"

	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"gorm.io/gorm"
)

// CropConfig mirrors the frontend CROPS array (indices 5-13).
type CropConfig struct {
	GrowTime   float64 // seconds [3]
	SeedCost   int     // [1]
	BaseYield  int     // [5]
	UnitSell   int     // [6]
	MinYield   int     // [7]
	MaxYield   int     // [8]
	DryRate    float64 // per hour [9]
	BugRate    float64 // per hour [10]
	WeedRate   float64 // per hour [11]
	MaxBug     int     // [12]
	MaxWeed    int     // [13]
}

// CROP_CONFIGS is indexed by crop_id (0-based), must stay in sync with frontend.
var CROP_CONFIGS = []CropConfig{
	{GrowTime: 12, SeedCost: 12, BaseYield: 4, UnitSell: 8, MinYield: 3, MaxYield: 5, DryRate: 0.06, BugRate: 0, WeedRate: 0.04, MaxBug: 0, MaxWeed: 1},      // 0 lettuce
	{GrowTime: 20, SeedCost: 20, BaseYield: 6, UnitSell: 10, MinYield: 4, MaxYield: 7, DryRate: 0.10, BugRate: 0.05, WeedRate: 0.05, MaxBug: 1, MaxWeed: 1},    // 1 pepper
	{GrowTime: 32, SeedCost: 35, BaseYield: 5, UnitSell: 19, MinYield: 3, MaxYield: 6, DryRate: 0.10, BugRate: 0.08, WeedRate: 0.08, MaxBug: 2, MaxWeed: 2},    // 2 eggplant
	{GrowTime: 48, SeedCost: 55, BaseYield: 8, UnitSell: 19, MinYield: 6, MaxYield: 10, DryRate: 0.14, BugRate: 0.09, WeedRate: 0.07, MaxBug: 2, MaxWeed: 2},   // 3 tomato
	{GrowTime: 70, SeedCost: 80, BaseYield: 12, UnitSell: 18, MinYield: 9, MaxYield: 14, DryRate: 0.16, BugRate: 0.12, WeedRate: 0.10, MaxBug: 2, MaxWeed: 2},  // 4 strawberry
	{GrowTime: 100, SeedCost: 120, BaseYield: 6, UnitSell: 57, MinYield: 4, MaxYield: 7, DryRate: 0.18, BugRate: 0.07, WeedRate: 0.12, MaxBug: 2, MaxWeed: 3}, // 5 corn
	{GrowTime: 135, SeedCost: 170, BaseYield: 4, UnitSell: 125, MinYield: 3, MaxYield: 5, DryRate: 0.13, BugRate: 0.04, WeedRate: 0.09, MaxBug: 1, MaxWeed: 2}, // 6 sunflower
	{GrowTime: 180, SeedCost: 240, BaseYield: 3, UnitSell: 240, MinYield: 2, MaxYield: 4, DryRate: 0.12, BugRate: 0.11, WeedRate: 0.15, MaxBug: 3, MaxWeed: 3}, // 7 pumpkin
	{GrowTime: 230, SeedCost: 320, BaseYield: 5, UnitSell: 196, MinYield: 3, MaxYield: 7, DryRate: 0.20, BugRate: 0.12, WeedRate: 0.14, MaxBug: 3, MaxWeed: 3}, // 8 watermelon
}

// Stage multipliers for event spawn rates.
// 0=seed, 1=sprout, 2=growing, 3=mature
var stageMultDry  = [4]float64{0.5, 1.0, 1.2, 0.0}
var stageMultBug  = [4]float64{0.0, 0.7, 1.2, 0.0}
var stageMultWeed = [4]float64{0.5, 1.0, 1.2, 0.0}

func getCropStageEnum(progress float64) int {
	if progress < 0.18 {
		return 0 // seed
	}
	if progress < 0.45 {
		return 1 // sprout
	}
	if progress < 0.90 {
		return 2 // growing
	}
	return 3 // mature
}

// ProcessFarm is the core engine: calculates all state changes since last_processed_at.
// Called before every action and on load.
func ProcessFarm(userID uint) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var plots []models.FarmPlot
	if err := database.DB.Where("user_id = ?", userID).Order("plot_index").Find(&plots).Error; err != nil {
		return fmt.Errorf("load plots: %w", err)
	}

	now := time.Now()
	needSave := false

	for i := range plots {
		p := &plots[i]
		if p.CropID == nil || p.Progress >= 1.0 {
			// Still update last_processed_at for next time
			if p.LastProcessedAt == nil || now.Sub(*p.LastProcessedAt) > 5*time.Second {
				p.LastProcessedAt = &now
				needSave = true
			}
			continue
		}

		var elapsed float64
		if p.LastProcessedAt != nil {
			elapsed = now.Sub(*p.LastProcessedAt).Seconds()
		} else {
			elapsed = 0
		}
		if elapsed <= 0 {
			continue
		}

		cid := *p.CropID
		if cid < 0 || cid >= len(CROP_CONFIGS) {
			continue
		}
		cfg := CROP_CONFIGS[cid]

		// Update game time
		user.GameTime += elapsed

		// ---- Apply growth ----
		stage := getCropStageEnum(p.Progress)
		if stage < 3 {
			speedMult := 1.0
			// Dry penalty
			if p.WaterState == 1 { // DRY
				speedMult *= 0.7
				p.DryTimer += elapsed
			}
			// Bug penalty
			if p.BugCount > 0 {
				penalty := 1.0 - float64(p.BugCount)*0.10
				if penalty < 0.3 {
					penalty = 0.3
				}
				speedMult *= penalty
			}
			// Weed penalty
			if p.WeedCount > 0 {
				penalty := 1.0 - float64(p.WeedCount)*0.05
				if penalty < 0.5 {
					penalty = 0.5
				}
				speedMult *= penalty
			}
			p.Progress += elapsed * speedMult / cfg.GrowTime
			if p.Progress > 1.0 {
				p.Progress = 1.0
			}
		}

		// ---- Spawn events (every 10-second check cycle) ----
		checkCycles := int(elapsed / 10.0)
		if checkCycles > 1000 {
			checkCycles = 1000 // cap for very long offline
		}
		stage = getCropStageEnum(p.Progress) // re-check after growth
		if stage < 3 {
			smDry := stageMultDry[stage]
			smBug := stageMultBug[stage]
			smWeed := stageMultWeed[stage]

			for j := 0; j < checkCycles; j++ {
				// Dry event
				if p.WaterState == 0 && now.Unix() >= int64(p.WaterProtectUntil) {
					if rand.Float64() < cfg.DryRate*(10.0/3600.0)*smDry {
						p.WaterState = 1
					}
				}
				// Bug event
				if p.BugCount < cfg.MaxBug && now.Unix() >= int64(p.BugProtectUntil) {
					if rand.Float64() < cfg.BugRate*(10.0/3600.0)*smBug {
						p.BugCount++
						if p.BugSince == 0 {
							p.BugSince = user.GameTime
						}
					}
				}
				// Weed event
				if p.WeedCount < cfg.MaxWeed && now.Unix() >= int64(p.WeedProtectUntil) {
					if rand.Float64() < cfg.WeedRate*(10.0/3600.0)*smWeed {
						p.WeedCount++
						if p.WeedSince == 0 {
							p.WeedSince = user.GameTime
						}
					}
				}
			}
		}

		// ---- Protection expiry ----
		if p.WaterState == 2 && user.GameTime >= p.WaterProtectUntil {
			p.WaterState = 0 // back to Normal
		}

		p.LastProcessedAt = &now
		needSave = true
	}

	if !needSave {
		return nil
	}

	// Batch save
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("save user: %w", err)
		}
		for i := range plots {
			if err := tx.Save(&plots[i]).Error; err != nil {
				return fmt.Errorf("save plot %d: %w", plots[i].PlotIndex, err)
			}
		}
		return nil
	})
}

// CalcHarvestYield computes the final yield for a given plot.
func CalcHarvestYield(p *models.FarmPlot, cfg CropConfig) int {
	base := cfg.BaseYield
	bonus := int(float64(base) * p.YieldBonusRate)
	loss := 0.0
	if p.DryTimer >= 5400 {
		loss += 0.10
	} else if p.DryTimer >= 1800 {
		loss += 0.05
	}
	if p.BugCount >= 3 {
		loss += 0.10
	} else if p.BugCount >= 2 {
		loss += 0.05
	}
	if p.WeedCount >= 3 {
		loss += 0.10
	} else if p.WeedCount >= 2 {
		loss += 0.05
	}
	if loss > 0.30 {
		loss = 0.30
	}
	result := base + bonus - int(float64(base)*loss)
	// Fert bonus guarantee: at least +1
	if p.YieldBonusRate > 0 && bonus == 0 {
		result++
	}
	if result < cfg.MinYield {
		result = cfg.MinYield
	}
	if result > cfg.MaxYield {
		result = cfg.MaxYield
	}
	return result
}

// ExecuteAction processes a player action with full server authority.
// 1. ProcessFarm (time-diff calculation)
// 2. Validate & execute the action
// 3. Return full farm state
func ExecuteAction(userID uint, req dto.ActionRequest) (*dto.ActionResponse, error) {
	// Step 1: Process elapsed time
	if err := ProcessFarm(userID); err != nil {
		return nil, fmt.Errorf("process farm: %w", err)
	}

	var message string
	var err error

	switch req.Action {
	case "plant":
		message, err = doPlant(userID, req)
	case "water":
		message, err = doWater(userID, req)
	case "fertilize":
		message, err = doFertilize(userID, req)
	case "harvest":
		message, err = doHarvest(userID, req)
	case "remove_bug":
		message, err = doRemoveBug(userID, req)
	case "remove_weed":
		message, err = doRemoveWeed(userID, req)
	case "shovel":
		message, err = doShovel(userID, req)
	case "shovel_all":
		message, err = doShovelAll(userID)
	case "harvest_all":
		message, err = doHarvestAll(userID)
	default:
		return nil, fmt.Errorf("unknown action: %s", req.Action)
	}
	if err != nil {
		return nil, err
	}

	// Step 3: Return full state
	loadResp, err := LoadFarm(userID)
	if err != nil {
		return nil, fmt.Errorf("load farm after action: %w", err)
	}
	return &dto.ActionResponse{LoadResponse: loadResp, Message: message}, nil
}

func getPlot(userID uint, plotIndex int) (*models.FarmPlot, error) {
	var p models.FarmPlot
	if err := database.DB.Where("user_id = ? AND plot_index = ?", userID, plotIndex).First(&p).Error; err != nil {
		return nil, fmt.Errorf("plot %d not found", plotIndex)
	}
	return &p, nil
}

func doPlant(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil || req.CropID == nil {
		return "", fmt.Errorf("plant requires plot_index and crop_id")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if !p.Unlocked {
		return "", fmt.Errorf("plot not unlocked")
	}
	if p.CropID != nil {
		return "", fmt.Errorf("plot already has a crop")
	}
	cid := *req.CropID
	if cid < 0 || cid >= len(CROP_CONFIGS) {
		return "", fmt.Errorf("invalid crop_id")
	}
	cfg := CROP_CONFIGS[cid]

	// Check gold
	var user models.User
	database.DB.First(&user, userID)
	if user.Gold < cfg.SeedCost {
		return "", fmt.Errorf("not enough gold (need %d)", cfg.SeedCost)
	}

	// Deduct gold and plant
	p.CropID = &cid
	p.Progress = 0
	p.WetTimer = 0
	p.WaterState = 0
	p.DryTimer = 0
	p.BugCount = 0
	p.BugSince = 0
	p.WeedCount = 0
	p.WeedSince = 0
	p.FertUsed = 0
	p.FertStageUsed = "{}"
	p.FertIDsUsed = "[]"
	p.YieldBonusRate = 0
	p.YieldLossRate = 0
	now := time.Now()
	p.LastProcessedAt = &now

	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Model(&models.User{}).Where("id = ?", userID).Update("gold", gorm.Expr("gold - ?", cfg.SeedCost))
		return tx.Save(p).Error
	})
	return fmt.Sprintf("种下了 %s!", cropNameZH(cid)), nil
}

func doWater(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("water requires plot_index")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.CropID == nil {
		return "", fmt.Errorf("no crop to water")
	}
	if p.Progress >= 1.0 {
		return "", fmt.Errorf("crop already mature")
	}
	if p.WaterState != 1 {
		return "", fmt.Errorf("crop is not thirsty")
	}

	cfg := CROP_CONFIGS[*p.CropID]
	protectDur := 40.0 * 60.0
	if cfg.GrowTime < 1800 {
		protectDur = 8.0 * 60.0
	} else if cfg.GrowTime < 5400 {
		protectDur = 20.0 * 60.0
	}
	var user models.User
	database.DB.First(&user, userID)

	p.WaterState = 2
	p.WaterProtectUntil = user.GameTime + protectDur
	p.DryTimer = 0
	p.WetTimer = 12

	database.DB.Save(p)
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("exp_val", gorm.Expr("exp_val + 1"))
	return "浇水成功!", nil
}

func doFertilize(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil || req.FertID == nil {
		return "", fmt.Errorf("fertilize requires plot_index and fert_id")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.CropID == nil {
		return "", fmt.Errorf("no crop to fertilize")
	}
	if p.Progress >= 1.0 {
		return "", fmt.Errorf("crop already mature, cannot fertilize")
	}
	fertID := *req.FertID
	if fertID < 0 || fertID >= len(CROP_CONFIGS) {
		// Fert IDs go beyond crop range, use frontend FERTILIZERS index
		// For now, return error for unknown fert
		return "", fmt.Errorf("invalid fertilizer id: %d", fertID)
	}
	if p.FertUsed >= 3 {
		return "", fmt.Errorf("already fertilized 3 times")
	}
	// TODO: check fert_stage_used, fert_ids_used, allowed_stages
	// For now, accept any valid fert ID
	cfg := CROP_CONFIGS[*p.CropID]
	_ = cfg // will be used for speed fertilizer effect

	// Apply effect based on fertID (0-6 matching frontend FERTILIZERS)
	switch fertID {
	case 0: // 初级速生肥
		reduction := cfg.GrowTime * 0.08
		if reduction > 600 { reduction = 600 }
		p.Progress += reduction / cfg.GrowTime
	case 1: // 中级速生肥
		reduction := cfg.GrowTime * 0.12
		if reduction > 1800 { reduction = 1800 }
		p.Progress += reduction / cfg.GrowTime
	case 2: // 高级速生肥
		reduction := cfg.GrowTime * 0.18
		if reduction > 3600 { reduction = 3600 }
		p.Progress += reduction / cfg.GrowTime
	case 3: // 保湿肥
		var user models.User
		database.DB.First(&user, userID)
		p.WaterProtectUntil = user.GameTime + 7200
		if p.WaterState == 1 {
			p.WaterState = 2
			p.DryTimer = 0
		}
	case 4: // 防虫肥
		var user models.User
		database.DB.First(&user, userID)
		p.BugProtectUntil = user.GameTime + 7200
	case 5: // 除草剂
		var user models.User
		database.DB.First(&user, userID)
		p.WeedProtectUntil = user.GameTime + 7200
	case 6: // 丰收肥
		p.YieldBonusRate += 0.10
	default:
		return "", fmt.Errorf("unknown fertilizer")
	}
	if p.Progress > 1.0 {
		p.Progress = 1.0
	}
	p.FertUsed++
	database.DB.Save(p)
	return "施肥成功!", nil
}

func doHarvest(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("harvest requires plot_index")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.CropID == nil || p.Progress < 1.0 {
		return "", fmt.Errorf("crop not mature")
	}
	cid := *p.CropID
	cfg := CROP_CONFIGS[cid]
	yieldCount := CalcHarvestYield(p, cfg)

	database.DB.Transaction(func(tx *gorm.DB) error {
		// Add to inventory
		var item models.InventoryItem
		if err := tx.Where("user_id = ? AND crop_id = ?", userID, cid).First(&item).Error; err != nil {
			item = models.InventoryItem{UserID: userID, CropID: uint(cid), Count: yieldCount}
			tx.Create(&item)
		} else {
			tx.Model(&item).Update("count", gorm.Expr("count + ?", yieldCount))
		}
		// Add exp
		tx.Model(&models.User{}).Where("id = ?", userID).Update("exp_val", gorm.Expr("exp_val + ?", cfg.BaseYield))
		// Reset plot
		p.CropID = nil
		p.Progress = 0
		p.WetTimer = 0
		p.WaterState = 0
		p.DryTimer = 0
		p.BugCount = 0
		p.BugSince = 0
		p.WeedCount = 0
		p.WeedSince = 0
		p.FertUsed = 0
		p.FertStageUsed = "{}"
		p.FertIDsUsed = "[]"
		p.YieldBonusRate = 0
		p.YieldLossRate = 0
		p.LandWork++
		if p.LandWork >= 30 && p.LandLevel < 4 {
			p.LandLevel++
			p.LandWork = 0
		}
		now := time.Now()
		p.LastProcessedAt = &now
		return tx.Save(p).Error
	})
	return fmt.Sprintf("收获 %s x%d!", cropNameZH(cid), yieldCount), nil
}

func doRemoveBug(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("remove_bug requires plot_index")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.BugCount <= 0 {
		return "", fmt.Errorf("no bugs here")
	}
	p.BugCount--
	if p.BugCount <= 0 {
		p.BugCount = 0
		p.BugSince = 0
		var user models.User
		database.DB.First(&user, userID)
		p.BugProtectUntil = user.GameTime + 120
	}
	database.DB.Save(p)
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("exp_val", gorm.Expr("exp_val + 1"))
	return "除虫成功!", nil
}

func doRemoveWeed(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("remove_weed requires plot_index")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.WeedCount <= 0 {
		return "", fmt.Errorf("no weeds here")
	}
	p.WeedCount--
	if p.WeedCount <= 0 {
		p.WeedCount = 0
		p.WeedSince = 0
		var user models.User
		database.DB.First(&user, userID)
		p.WeedProtectUntil = user.GameTime + 120
	}
	database.DB.Save(p)
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("exp_val", gorm.Expr("exp_val + 1"))
	return "除草成功!", nil
}

func doShovel(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("shovel requires plot_index")
	}
	p, err := getPlot(userID, *req.PlotIndex)
	if err != nil {
		return "", err
	}
	if p.CropID == nil {
		return "", fmt.Errorf("no crop to shovel")
	}
	name := cropNameZH(*p.CropID)
	p.CropID = nil
	p.Progress = 0
	p.WetTimer = 0
	p.WaterState = 0
	p.DryTimer = 0
	p.BugCount = 0
	p.WeedCount = 0
	p.FertUsed = 0
	p.YieldBonusRate = 0
	now := time.Now()
	p.LastProcessedAt = &now
	database.DB.Save(p)
	return fmt.Sprintf("铲除了 %s", name), nil
}

func doHarvestAll(userID uint) (string, error) {
	var plots []models.FarmPlot
	database.DB.Where("user_id = ? AND crop_id IS NOT NULL AND progress >= 1.0", userID).Find(&plots)
	if len(plots) == 0 {
		return "", fmt.Errorf("no mature crops")
	}
	count := 0
	for _, p := range plots {
		cid := *p.CropID
		cfg := CROP_CONFIGS[cid]
		yieldCount := CalcHarvestYield(&p, cfg)
		database.DB.Transaction(func(tx *gorm.DB) error {
			var item models.InventoryItem
			if err := tx.Where("user_id = ? AND crop_id = ?", userID, cid).First(&item).Error; err != nil {
				item = models.InventoryItem{UserID: userID, CropID: uint(cid), Count: yieldCount}
				tx.Create(&item)
			} else {
				tx.Model(&item).Update("count", gorm.Expr("count + ?", yieldCount))
			}
			p.CropID = nil
			p.Progress = 0
			p.WetTimer = 0
			p.WaterState = 0
			p.DryTimer = 0
			p.BugCount = 0
			p.WeedCount = 0
			p.FertUsed = 0
			p.YieldBonusRate = 0
			p.LandWork++
			if p.LandWork >= 30 && p.LandLevel < 4 {
				p.LandLevel++
				p.LandWork = 0
			}
			now := time.Now()
			p.LastProcessedAt = &now
			return tx.Save(&p).Error
		})
		count++
	}
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("exp_val", gorm.Expr("exp_val + ?", count*5))
	return fmt.Sprintf("一键收获了 %d 个作物!", count), nil
}

func doShovelAll(userID uint) (string, error) {
	var plots []models.FarmPlot
	database.DB.Where("user_id = ? AND crop_id IS NOT NULL", userID).Find(&plots)
	if len(plots) == 0 {
		return "", fmt.Errorf("no crops to shovel")
	}
	count := 0
	for _, p := range plots {
		p.CropID = nil
		p.Progress = 0
		p.WetTimer = 0
		p.WaterState = 0
		p.DryTimer = 0
		p.BugCount = 0
		p.WeedCount = 0
		p.FertUsed = 0
		p.YieldBonusRate = 0
		now := time.Now()
		p.LastProcessedAt = &now
		database.DB.Save(&p)
		count++
	}
	return fmt.Sprintf("铲除了全部 %d 个作物!", count), nil
}

// cropNameZH returns the Chinese name for a crop ID.
func cropNameZH(cid int) string {
	names := []string{"生菜", "辣椒", "茄子", "西红柿", "草莓", "玉米", "向日葵", "南瓜", "西瓜"}
	if cid >= 0 && cid < len(names) {
		return names[cid]
	}
	return "未知作物"
}

