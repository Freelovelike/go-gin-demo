package services

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"gorm.io/gorm"
)

// CropConfig mirrors the frontend CROPS array (indices 5-13).
type CropConfig struct {
	Name      string  // [0] Chinese display name
	Key       string  // [4] texture key, e.g. "tomato"
	GrowTime  float64 // seconds [3]
	SeedCost  int     // [1]
	BaseYield int     // [5]
	UnitSell  int     // [6]
	MinYield  int     // [7]
	MaxYield  int     // [8]
	DryRate   float64 // per hour [9]
	BugRate   float64 // per hour [10]
	WeedRate  float64 // per hour [11]
	MaxBug    int     // [12]
	MaxWeed   int     // [13]
}

type FertilizerConfig struct {
	Name            string
	Cost            int
	Type            string
	EffectValue     float64
	AllowedStages   []int
	PerCropLimit    int
	MaxMinutesLimit int
}

// CROP_CONFIGS is indexed by crop_id (0-based), must stay in sync with frontend.
var CROP_CONFIGS = []CropConfig{
	{Name: "生菜", Key: "lettuce", GrowTime: 12, SeedCost: 12, BaseYield: 4, UnitSell: 8, MinYield: 3, MaxYield: 5, DryRate: 0.06, BugRate: 0, WeedRate: 0.04, MaxBug: 0, MaxWeed: 1},           // 0 lettuce
	{Name: "辣椒", Key: "pepper", GrowTime: 20, SeedCost: 20, BaseYield: 6, UnitSell: 10, MinYield: 4, MaxYield: 7, DryRate: 0.10, BugRate: 0.05, WeedRate: 0.05, MaxBug: 1, MaxWeed: 1},        // 1 pepper
	{Name: "茄子", Key: "eggplant", GrowTime: 32, SeedCost: 35, BaseYield: 5, UnitSell: 19, MinYield: 3, MaxYield: 6, DryRate: 0.10, BugRate: 0.08, WeedRate: 0.08, MaxBug: 2, MaxWeed: 2},      // 2 eggplant
	{Name: "西红柿", Key: "tomato", GrowTime: 48, SeedCost: 55, BaseYield: 8, UnitSell: 19, MinYield: 6, MaxYield: 10, DryRate: 0.14, BugRate: 0.09, WeedRate: 0.07, MaxBug: 2, MaxWeed: 2},      // 3 tomato
	{Name: "草莓", Key: "strawberry", GrowTime: 70, SeedCost: 80, BaseYield: 12, UnitSell: 18, MinYield: 9, MaxYield: 14, DryRate: 0.16, BugRate: 0.12, WeedRate: 0.10, MaxBug: 2, MaxWeed: 2},  // 4 strawberry
	{Name: "玉米", Key: "corn", GrowTime: 100, SeedCost: 120, BaseYield: 6, UnitSell: 57, MinYield: 4, MaxYield: 7, DryRate: 0.18, BugRate: 0.07, WeedRate: 0.12, MaxBug: 2, MaxWeed: 3},        // 5 corn
	{Name: "向日葵", Key: "sunflower", GrowTime: 135, SeedCost: 170, BaseYield: 4, UnitSell: 125, MinYield: 3, MaxYield: 5, DryRate: 0.13, BugRate: 0.04, WeedRate: 0.09, MaxBug: 1, MaxWeed: 2}, // 6 sunflower
	{Name: "南瓜", Key: "pumpkin", GrowTime: 180, SeedCost: 240, BaseYield: 3, UnitSell: 240, MinYield: 2, MaxYield: 4, DryRate: 0.12, BugRate: 0.11, WeedRate: 0.15, MaxBug: 3, MaxWeed: 3},    // 7 pumpkin
	{Name: "西瓜", Key: "watermelon", GrowTime: 230, SeedCost: 320, BaseYield: 5, UnitSell: 196, MinYield: 3, MaxYield: 7, DryRate: 0.20, BugRate: 0.12, WeedRate: 0.14, MaxBug: 3, MaxWeed: 3}, // 8 watermelon
}

// FERTILIZER_COSTS is indexed by fertilizer_id (0-6), must match frontend FERTILIZERS[*][1].
var FERTILIZER_COSTS = []int{15, 40, 80, 30, 35, 25, 60}

const FERTILIZER_COUNT = 7

var FERTILIZER_CONFIGS = []FertilizerConfig{
	{Name: "初级速生肥", Cost: 15, Type: "speed", EffectValue: 0.08, AllowedStages: []int{0, 1}, PerCropLimit: 1, MaxMinutesLimit: 10},
	{Name: "中级速生肥", Cost: 40, Type: "speed", EffectValue: 0.12, AllowedStages: []int{1, 2}, PerCropLimit: 1, MaxMinutesLimit: 30},
	{Name: "高级速生肥", Cost: 80, Type: "speed", EffectValue: 0.18, AllowedStages: []int{2}, PerCropLimit: 1, MaxMinutesLimit: 60},
	{Name: "保湿肥", Cost: 30, Type: "water_protect", EffectValue: 7200.0, AllowedStages: []int{0, 1, 2}, PerCropLimit: 1, MaxMinutesLimit: 0},
	{Name: "防虫肥", Cost: 35, Type: "bug_protect", EffectValue: 7200.0, AllowedStages: []int{1, 2}, PerCropLimit: 1, MaxMinutesLimit: 0},
	{Name: "除草剂", Cost: 25, Type: "weed_protect", EffectValue: 7200.0, AllowedStages: []int{-1, 0, 1, 2}, PerCropLimit: 1, MaxMinutesLimit: 0},
	{Name: "丰收肥", Cost: 60, Type: "yield_bonus", EffectValue: 0.10, AllowedStages: []int{2}, PerCropLimit: 1, MaxMinutesLimit: 0},
}

// Stage multipliers for event spawn rates.
// 0=seed, 1=sprout, 2=growing, 3=mature
var stageMultDry = [4]float64{0.5, 1.0, 1.2, 0.0}
var stageMultBug = [4]float64{0.0, 0.7, 1.2, 0.0}
var stageMultWeed = [4]float64{0.5, 1.0, 1.2, 0.0}

const (
	stageSeedEnd      = 0.18
	stageSproutEnd    = 0.45
	stageGrowingEnd   = 0.90
	renderStageMidEnd = 0.72
)

func getCropStageEnum(progress float64) int {
	if progress < stageSeedEnd {
		return 0 // seed
	}
	if progress < stageSproutEnd {
		return 1 // sprout
	}
	if progress < stageGrowingEnd {
		return 2 // growing
	}
	return 3 // mature
}

func intInSlice(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func countFertilizerID(values []int, target int) int {
	count := 0
	for _, value := range values {
		if value == target {
			count++
		}
	}
	return count
}

// eventProbability returns the chance of at least one event occurring across
// `cycles` 10-second check windows, given a per-hour `rate` and a stage
// multiplier `mult`. Equivalent to the old per-cycle Bernoulli loop but in
// closed form: 1 - (1-p)^cycles, with no iteration and no offline cap.
func eventProbability(rate, mult, cycles float64) float64 {
	if rate <= 0 || mult <= 0 || cycles <= 0 {
		return 0
	}
	pPerCycle := rate * (10.0 / 3600.0) * mult
	if pPerCycle <= 0 {
		return 0
	}
	if pPerCycle >= 1 {
		return 1
	}
	return 1 - math.Pow(1-pPerCycle, cycles)
}

func unixNow(t time.Time) float64 {
	return float64(t.Unix())
}

func unixSeconds(t *time.Time) float64 {
	if t == nil {
		return 0
	}
	return float64(t.UnixMilli()) / 1000.0
}

func currentGrowthSpeedMultiplier(p models.FarmPlot) float64 {
	speedMult := 1.0
	if p.WaterState == 1 {
		speedMult *= 0.7
	}
	if p.BugCount > 0 {
		penalty := 1.0 - float64(p.BugCount)*0.10
		if penalty < 0.3 {
			penalty = 0.3
		}
		speedMult *= penalty
	}
	if p.WeedCount > 0 {
		penalty := 1.0 - float64(p.WeedCount)*0.05
		if penalty < 0.5 {
			penalty = 0.5
		}
		speedMult *= penalty
	}
	return speedMult
}

func estimateMatureAt(p models.FarmPlot, now time.Time) float64 {
	if p.CropID == nil || p.Progress >= 1.0 {
		return 0
	}
	cid := *p.CropID
	if cid < 0 || cid >= len(CROP_CONFIGS) {
		return 0
	}
	speedMult := currentGrowthSpeedMultiplier(p)
	if speedMult <= 0 {
		return 0
	}
	remaining := (1.0 - p.Progress) * CROP_CONFIGS[cid].GrowTime / speedMult
	return unixNow(now) + remaining
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
	maxElapsed := 0.0

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
		if elapsed > maxElapsed {
			maxElapsed = elapsed
		}

		cid := *p.CropID
		if cid < 0 || cid >= len(CROP_CONFIGS) {
			continue
		}
		cfg := CROP_CONFIGS[cid]

		// ---- Apply growth ----
		stage := getCropStageEnum(p.Progress)
		if stage < 3 {
			if p.WaterState == 1 { // DRY
				p.DryTimer += elapsed
			}
			speedMult := currentGrowthSpeedMultiplier(*p)
			p.Progress += elapsed * speedMult / cfg.GrowTime
			if p.Progress > 1.0 {
				p.Progress = 1.0
			}
		}

		// ---- Spawn events (closed-form over the elapsed window) ----
		// Each event was originally Bernoulli-sampled once per 10s cycle. Sampling
		// in a loop is O(elapsed) and was capped at 1000 cycles, which silently
		// changed the rules past ~2.8h offline. Instead, compute the probability
		// of at least one event across all cycles in closed form:
		//   pAtLeastOne = 1 - (1 - pPerCycle)^cycles
		// and draw once (dry) or once per open slot (bug/weed).
		cycles := elapsed / 10.0
		stage = getCropStageEnum(p.Progress) // re-check after growth
		if stage < 3 {
			smDry := stageMultDry[stage]
			smBug := stageMultBug[stage]
			smWeed := stageMultWeed[stage]

			// Dry: single binary state.
			if p.WaterState == 0 && unixNow(now) >= p.WaterProtectUntil {
				if rand.Float64() < eventProbability(cfg.DryRate, smDry, cycles) {
					p.WaterState = 1
				}
			}
			// Bug: roll each remaining slot up to MaxBug.
			if p.BugCount < cfg.MaxBug && unixNow(now) >= p.BugProtectUntil {
				prob := eventProbability(cfg.BugRate, smBug, cycles)
				for p.BugCount < cfg.MaxBug && rand.Float64() < prob {
					p.BugCount++
					if p.BugSince == 0 {
						p.BugSince = user.GameTime
					}
				}
			}
			// Weed: roll each remaining slot up to MaxWeed.
			if p.WeedCount < cfg.MaxWeed && unixNow(now) >= p.WeedProtectUntil {
				prob := eventProbability(cfg.WeedRate, smWeed, cycles)
				for p.WeedCount < cfg.MaxWeed && rand.Float64() < prob {
					p.WeedCount++
					if p.WeedSince == 0 {
						p.WeedSince = user.GameTime
					}
				}
			}
		}

		// ---- Protection expiry ----
		if p.WaterState == 2 && unixNow(now) >= p.WaterProtectUntil {
			p.WaterState = 0 // back to Normal
		}

		p.LastProcessedAt = &now
		needSave = true
	}

	if !needSave {
		return nil
	}
	user.GameTime += maxElapsed

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

// checkLevelUp processes level-ups after exp gains.
func checkLevelUp(userID uint) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return
	}
	changed := false
	for user.ExpVal >= user.ExpToLvl {
		user.ExpVal -= user.ExpToLvl
		user.Level++
		user.ExpToLvl = int(float64(user.ExpToLvl) * 1.5)
		changed = true
	}
	if changed {
		database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"exp_val":    user.ExpVal,
			"level":      user.Level,
			"exp_to_lvl": user.ExpToLvl,
		})
	}
}

func awardExp(userID uint, amount int) error {
	if amount <= 0 {
		return nil
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("exp_val", gorm.Expr("exp_val + ?", amount)).Error; err != nil {
		return err
	}
	checkLevelUp(userID)
	return nil
}

// ExecuteAction processes a player action with full server authority.
// 1. ProcessFarm (time-diff calculation)
// 2. Validate & execute the action
// 3. Return full farm state
func ExecuteAction(userID uint, req dto.ActionRequest) (*dto.ActionResponse, error) {
	// Serialize all writes for this user — the client fires actions from
	// several parallel HTTPRequest nodes, so without this two requests could
	// both pass a gold check before either deducts.
	defer lockUser(userID)()

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
	case "sell_all":
		message, err = doSellAll(userID)
	case "sell":
		message, err = doSell(userID, req)
	case "buy_fertilizer":
		message, err = doBuyFertilizer(userID, req)
	case "reclaim":
		message, err = doReclaim(userID, req)
	default:
		return nil, fmt.Errorf("unknown action: %s", req.Action)
	}
	if err != nil {
		return nil, err
	}

	// Step 3: Return full state
	loadResp, err := loadFarmLocked(userID)
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
	p.PlantedAt = &now
	p.LastProcessedAt = &now

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Model(&models.User{}).Where("id = ?", userID).Update("gold", gorm.Expr("gold - ?", cfg.SeedCost))
		return tx.Save(p).Error
	})
	if err != nil {
		return "", err
	}
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
	p.WaterState = 2
	p.WaterProtectUntil = unixNow(time.Now()) + protectDur
	p.DryTimer = 0
	p.WetTimer = 12

	if err := database.DB.Save(p).Error; err != nil {
		return "", err
	}
	if err := awardExp(userID, 1); err != nil {
		return "", err
	}
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
	if fertID < 0 || fertID >= FERTILIZER_COUNT {
		return "", fmt.Errorf("invalid fertilizer id: %d", fertID)
	}
	// Check and deduct from fertilizer inventory
	var fi models.FertilizerInventory
	if err := database.DB.Where("user_id = ? AND fert_index = ?", userID, fertID).First(&fi).Error; err != nil || fi.Count <= 0 {
		return "", fmt.Errorf("没有该肥料，请先购买")
	}
	if p.FertUsed >= 3 {
		return "", fmt.Errorf("already fertilized 3 times")
	}
	cropCfg := CROP_CONFIGS[*p.CropID]
	fertCfg := FERTILIZER_CONFIGS[fertID]
	stage := getCropStageEnum(p.Progress)
	if !intInSlice(fertCfg.AllowedStages, stage) {
		return "", fmt.Errorf("该肥料不能在当前阶段使用")
	}
	usedIDs := parseFertIDsUsed(p.FertIDsUsed)
	if fertCfg.PerCropLimit > 0 && countFertilizerID(usedIDs, fertID) >= fertCfg.PerCropLimit {
		return "", fmt.Errorf("该作物已使用过这种肥料")
	}
	stageUsed := parseFertStageUsed(p.FertStageUsed)

	switch fertCfg.Type {
	case "speed":
		reduction := cropCfg.GrowTime * fertCfg.EffectValue
		if fertCfg.MaxMinutesLimit > 0 {
			maxReduction := float64(fertCfg.MaxMinutesLimit) * 60.0
			if reduction > maxReduction {
				reduction = maxReduction
			}
		}
		p.Progress += reduction / cropCfg.GrowTime
	case "water_protect":
		p.WaterProtectUntil = unixNow(time.Now()) + fertCfg.EffectValue
		if p.WaterState == 1 {
			p.WaterState = 2
			p.DryTimer = 0
		}
	case "bug_protect":
		p.BugProtectUntil = unixNow(time.Now()) + fertCfg.EffectValue
	case "weed_protect":
		p.WeedProtectUntil = unixNow(time.Now()) + fertCfg.EffectValue
	case "yield_bonus":
		p.YieldBonusRate += fertCfg.EffectValue
	default:
		return "", fmt.Errorf("unknown fertilizer type: %s", fertCfg.Type)
	}
	if p.Progress > 1.0 {
		p.Progress = 1.0
	}
	p.FertUsed++
	stageKey := fmt.Sprintf("%d", stage)
	stageUsed[stageKey] = stageUsed[stageKey] + 1
	usedIDs = append(usedIDs, fertID)
	p.FertStageUsed = encodeFertStageUsed(stageUsed)
	p.FertIDsUsed = encodeFertIDsUsed(usedIDs)
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(p).Error; err != nil {
			return err
		}
		return tx.Model(&fi).Update("count", gorm.Expr("count - 1")).Error
	}); err != nil {
		return "", err
	}
	return "施肥成功!", nil
}

// Land upgrade rules — must match the frontend constants.
const (
	landUpgradeWorkRequired = 30
	landLevelMax            = 4
)

// clearPlot resets a plot's crop, care, and fertilizer state. Shared by both
// harvest and shovel paths so the (long) reset stays in one place.
func clearPlot(p *models.FarmPlot) {
	p.CropID = nil
	p.PlantedAt = nil
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
}

// applyLandWork credits one unit of farming work and upgrades the land level
// when the threshold is reached. (B5: single source for the upgrade rule.)
func applyLandWork(p *models.FarmPlot) {
	p.LandWork++
	if p.LandWork >= landUpgradeWorkRequired && p.LandLevel < landLevelMax {
		p.LandLevel++
		p.LandWork = 0
	}
}

// addToInventory upserts the inventory count for a crop within a transaction.
func addToInventory(tx *gorm.DB, userID uint, cid, amount int) error {
	var item models.InventoryItem
	if err := tx.Where("user_id = ? AND crop_id = ?", userID, cid).First(&item).Error; err != nil {
		item = models.InventoryItem{UserID: userID, CropID: uint(cid), Count: amount}
		return tx.Create(&item).Error
	}
	return tx.Model(&item).Update("count", gorm.Expr("count + ?", amount)).Error
}

// harvestOnePlot harvests a single (assumed-mature) plot inside a transaction:
// credits inventory, resets the plot, and applies land work. Returns the yield.
// Exp is awarded by the caller because single vs. bulk harvest use different
// exp formulas.
func harvestOnePlot(tx *gorm.DB, userID uint, p *models.FarmPlot) (int, error) {
	cid := *p.CropID
	yieldCount := CalcHarvestYield(p, CROP_CONFIGS[cid])
	if err := addToInventory(tx, userID, cid, yieldCount); err != nil {
		return 0, err
	}
	clearPlot(p)
	applyLandWork(p)
	if err := tx.Save(p).Error; err != nil {
		return 0, err
	}
	return yieldCount, nil
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

	var yieldCount int
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var e error
		yieldCount, e = harvestOnePlot(tx, userID, p)
		if e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if err := awardExp(userID, CROP_CONFIGS[cid].BaseYield); err != nil {
		return "", err
	}
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
		p.BugProtectUntil = unixNow(time.Now()) + 120
	}
	if err := database.DB.Save(p).Error; err != nil {
		return "", err
	}
	if err := awardExp(userID, 1); err != nil {
		return "", err
	}
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
		p.WeedProtectUntil = unixNow(time.Now()) + 120
	}
	if err := database.DB.Save(p).Error; err != nil {
		return "", err
	}
	if err := awardExp(userID, 1); err != nil {
		return "", err
	}
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
	clearPlot(p)
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
	// Single transaction over all plots: previously each plot ran its own
	// transaction, so a mid-loop failure left a half-harvested farm.
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for i := range plots {
			if _, e := harvestOnePlot(tx, userID, &plots[i]); e != nil {
				return e
			}
			count++
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if err := awardExp(userID, count*5); err != nil {
		return "", err
	}
	return fmt.Sprintf("一键收获了 %d 个作物!", count), nil
}

func doShovelAll(userID uint) (string, error) {
	var plots []models.FarmPlot
	database.DB.Where("user_id = ? AND crop_id IS NOT NULL", userID).Find(&plots)
	if len(plots) == 0 {
		return "", fmt.Errorf("no crops to shovel")
	}
	count := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for i := range plots {
			clearPlot(&plots[i])
			if e := tx.Save(&plots[i]).Error; e != nil {
				return e
			}
			count++
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("铲除了全部 %d 个作物!", count), nil
}

// cropNameZH returns the Chinese name for a crop ID (single source: CROP_CONFIGS).
func cropNameZH(cid int) string {
	if cid >= 0 && cid < len(CROP_CONFIGS) {
		return CROP_CONFIGS[cid].Name
	}
	return "未知作物"
}

func doSell(userID uint, req dto.ActionRequest) (string, error) {
	if req.CropID == nil || req.Count == nil {
		return "", fmt.Errorf("sell requires crop_id and count")
	}
	resp, err := sellCropLocked(userID, *req.CropID, *req.Count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("售出 %s x%d，获得 %d 金币", cropNameZH(*req.CropID), resp.SoldCount, resp.GoldEarned), nil
}

// doSellAll sells all inventory items server-side.
func doSellAll(userID uint) (string, error) {
	var items []models.InventoryItem
	if err := database.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return "", fmt.Errorf("load inventory: %w", err)
	}
	totalCount := 0
	totalGold := 0
	for _, item := range items {
		if item.Count <= 0 {
			continue
		}
		cid := int(item.CropID)
		if cid < 0 || cid >= len(CROP_CONFIGS) {
			continue
		}
		totalCount += item.Count
		totalGold += CROP_CONFIGS[cid].UnitSell * item.Count
	}
	if totalCount == 0 {
		return "", fmt.Errorf("背包是空的")
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Clear inventory
		if err := tx.Where("user_id = ?", userID).Delete(&models.InventoryItem{}).Error; err != nil {
			return err
		}
		// Add gold
		return tx.Model(&models.User{}).Where("id = ?", userID).
			Update("gold", gorm.Expr("gold + ?", totalGold)).Error
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("全部卖出 %d 个作物，获得 %d 金币", totalCount, totalGold), nil
}

// doBuyFertilizer purchases a fertilizer and adds it to inventory.
func doBuyFertilizer(userID uint, req dto.ActionRequest) (string, error) {
	if req.FertID == nil {
		return "", fmt.Errorf("buy_fertilizer requires fert_id")
	}
	fertID := *req.FertID
	if fertID < 0 || fertID >= FERTILIZER_COUNT {
		return "", fmt.Errorf("invalid fertilizer id: %d", fertID)
	}
	cost := FERTILIZER_COSTS[fertID]
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}
	if user.Gold < cost {
		return "", fmt.Errorf("金币不足 (need %d)", cost)
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Deduct gold
		if err := tx.Model(&models.User{}).Where("id = ?", userID).
			Update("gold", gorm.Expr("gold - ?", cost)).Error; err != nil {
			return err
		}
		// Upsert fertilizer inventory
		var fi models.FertilizerInventory
		if err := tx.Where("user_id = ? AND fert_index = ?", userID, fertID).First(&fi).Error; err != nil {
			fi = models.FertilizerInventory{UserID: userID, FertIndex: fertID, Count: 1}
			return tx.Create(&fi).Error
		}
		return tx.Model(&fi).Update("count", gorm.Expr("count + 1")).Error
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("购买 %s 成功!", FERTILIZER_CONFIGS[fertID].Name), nil
}

// Reclaim constants — must match frontend formulas.
const (
	baseReclaimCost = 60
	reclaimCostStep = 35
	totalPlots      = 30
)

// doReclaim unlocks a plot with server-authoritative gold/level/order checks.
func doReclaim(userID uint, req dto.ActionRequest) (string, error) {
	if req.PlotIndex == nil {
		return "", fmt.Errorf("reclaim requires plot_index")
	}
	idx := *req.PlotIndex
	if idx < 0 || idx >= totalPlots {
		return "", fmt.Errorf("invalid plot_index")
	}

	p, err := getPlot(userID, idx)
	if err != nil {
		return "", err
	}
	if p.Unlocked {
		return "", fmt.Errorf("plot already unlocked")
	}

	// Enforce sequential unlock: the plot must be the first locked one.
	var firstLocked models.FarmPlot
	if err := database.DB.Where("user_id = ? AND unlocked = ?", userID, false).
		Order("plot_index").First(&firstLocked).Error; err != nil {
		return "", fmt.Errorf("no locked plot available")
	}
	if firstLocked.PlotIndex != idx {
		return "", fmt.Errorf("请按顺序先开垦下一块土地")
	}

	requiredLevel := idx + 1
	cost := baseReclaimCost + idx*reclaimCostStep

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}
	if user.Level < requiredLevel {
		return "", fmt.Errorf("等级不足! 这块地需要等级 %d", requiredLevel)
	}
	if user.Gold < cost {
		return "", fmt.Errorf("金币不足! 开垦需要 %d 金币", cost)
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).
			Update("gold", gorm.Expr("gold - ?", cost)).Error; err != nil {
			return err
		}
		p.Unlocked = true
		p.LandLevel = 1
		p.LandWork = 0
		p.CropID = nil
		p.Progress = 0
		p.WetTimer = 0
		now := time.Now()
		p.LastProcessedAt = &now
		return tx.Save(p).Error
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("开垦成功! 解锁第 %d 块地", idx+1), nil
}
