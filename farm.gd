extends Node2D

const CropAtlas = preload("res://scripts/crop_atlas.gd")
const TOOL_ICON_TEXTURES: Array[Texture2D] = [
	preload("res://assets/ui/icons/tool_select.png"),
	preload("res://assets/ui/icons/tool_water.png"),
	preload("res://assets/ui/icons/tool_fertilizer.png"),
	preload("res://assets/ui/icons/tool_harvest.png"),
	preload("res://assets/ui/icons/tool_shovel.png"),
]
const LAND_TEXTURE_PATHS := {
	"locked": "res://assets/land/land_grass_locked.png",
	"yellow_dry": "res://assets/land/land_yellow_dry.png",
	"yellow_wet": "res://assets/land/land_yellow_wet.png",
	"red_dry": "res://assets/land/land_red_dry.png",
	"red_wet": "res://assets/land/land_red_wet.png",
	"black_dry": "res://assets/land/land_black_dry.png",
	"black_wet": "res://assets/land/land_black_wet.png",
}
var land_textures: Dictionary = {}
var land_texture_source_rects: Dictionary = {}

# ===================== Iso Farm 2.5D =====================
const VIEW_W := 1448.0
const VIEW_H := 1086.0
const TW := 168
const TH := 84
const COLS := 6
const ROWS := 5
const OX := 500.0
const OY := 320.0
const SAVE_PATH := "user://qq_farm_save.json"
const PLOT_ANCHORS_PATH := "PlotAnchors"
const INITIAL_UNLOCKED_PLOTS := 1
const BASE_RECLAIM_COST := 60
const RECLAIM_COST_STEP := 35
const LAND_LEVEL_LOCKED := 0
const LAND_LEVEL_MAX := 4
const LAND_UPGRADE_WORK_REQUIRED := 30

var CROPS: Array = []
var CROP_COLORS: Array = []

var gold := 200
var level := 1
var exp_val := 0
var exp_to_level := 100

var farm: Array = []
var selected_seed := -1
var shop_open := false
var hover_col := -1
var hover_row := -1
var mouse_held := false
var last_action_col := -1
var last_action_row := -1
# 工具模式: 0=普通 1=浇水 2=施肥 3=收获 4=铲除
var tool_mode := 0

# 背包系统
var inventory = {}
var inventory_open := false
var side_panel_open := true

var toast_text := ""
var toast_timer := 0.0
var save_timer := 0.0

func _ready():
	CROPS = [
		["生菜", 12, 32, 12, "lettuce"],
		["辣椒", 20, 58, 20, "pepper"],
		["茄子", 35, 95, 32, "eggplant"],
		["西红柿", 55, 150, 48, "tomato"],
		["草莓", 80, 220, 70, "strawberry"],
		["玉米", 120, 340, 100, "corn"],
		["向日葵", 170, 500, 135, "sunflower"],
		["南瓜", 240, 720, 180, "pumpkin"],
		["西瓜", 320, 980, 230, "watermelon"],
	]
	CROP_COLORS = [
		[Color(0.42, 0.76, 0.34), Color(0.72, 0.96, 0.46)],
		[Color(0.24, 0.68, 0.22), Color(0.92, 0.24, 0.18)],
		[Color(0.30, 0.72, 0.26), Color(0.48, 0.22, 0.62)],
		[Color(0.28, 0.76, 0.28), Color(0.96, 0.24, 0.18)],
		[Color(0.28, 0.76, 0.34), Color(0.98, 0.18, 0.26)],
		[Color(0.36, 0.76, 0.22), Color(0.98, 0.86, 0.20)],
		[Color(0.28, 0.64, 0.24), Color(0.98, 0.74, 0.16)],
		[Color(0.30, 0.70, 0.24), Color(0.96, 0.52, 0.12)],
		[Color(0.26, 0.66, 0.22), Color(0.34, 0.86, 0.28)],
	]
	farm = []
	for _r in range(ROWS):
		var row: Array = []
		for _c in range(COLS):
			row.append(_create_empty_cell(_c, _r))
		farm.append(row)
	_load_land_textures()
	_load_game()

func _load_land_textures():
	land_textures.clear()
	land_texture_source_rects.clear()
	for key in LAND_TEXTURE_PATHS.keys():
		var texture := load(str(LAND_TEXTURE_PATHS[key])) as Texture2D
		land_textures[key] = texture
		if texture != null:
			land_texture_source_rects[key] = _get_texture_alpha_bounds(texture)

func iso2screen(c: int, r: int) -> Vector2:
	return Vector2(OX + (c - r) * TW * 0.5, OY + (c + r) * TH * 0.5)

func _get_plot_position(c: int, r: int) -> Vector2:
	var anchors := get_node_or_null(PLOT_ANCHORS_PATH)
	if anchors != null:
		var plot := anchors.get_node_or_null("Plot_%d_%d" % [r, c])
		if plot is Node2D:
			return (plot as Node2D).global_position
	return iso2screen(c, r)

func _create_empty_cell(col: int, row: int) -> Dictionary:
	var initial_land_level := 1 if _get_plot_index(col, row) < INITIAL_UNLOCKED_PLOTS else LAND_LEVEL_LOCKED
	return {
		"crop_id": -1,
		"progress": 0.0,
		"wet_timer": 0.0,
		"unlocked": initial_land_level > LAND_LEVEL_LOCKED,
		"land_level": initial_land_level,
		"land_work": 0,
	}

func _get_plot_index(col: int, row: int) -> int:
	return row * COLS + col

func _get_reclaim_level(col: int, row: int) -> int:
	return _get_plot_index(col, row) + 1

func _get_reclaim_cost(col: int, row: int) -> int:
	return BASE_RECLAIM_COST + _get_plot_index(col, row) * RECLAIM_COST_STEP

func _is_cell_unlocked(cell: Dictionary) -> bool:
	return int(cell.get("land_level", LAND_LEVEL_LOCKED)) > LAND_LEVEL_LOCKED

func _get_land_level_name(land_level: int) -> String:
	if land_level <= LAND_LEVEL_LOCKED:
		return "未开垦"
	if land_level == 1:
		return "黄土地Lv1"
	if land_level == 2:
		return "黄土地Lv2"
	if land_level == 3:
		return "红土地"
	return "黑土地"

func screen2iso(pos: Vector2) -> Vector2i:
	var dx := (pos.x - OX) / (TW * 0.5)
	var dy := (pos.y - OY) / (TH * 0.5)
	return Vector2i(int((dx + dy) * 0.5), int((dy - dx) * 0.5))

func iso_corners(cx: float, cy: float) -> PackedVector2Array:
	return PackedVector2Array([
		Vector2(cx, cy - TH * 0.5),
		Vector2(cx + TW * 0.5, cy),
		Vector2(cx, cy + TH * 0.5),
		Vector2(cx - TW * 0.5, cy),
	])

func in_diamond(px: float, py: float, cx: float, cy: float) -> bool:
	return absf(px - cx) / (TW * 0.5) + absf(py - cy) / (TH * 0.5) <= 1.0

func _process(delta: float):
	for r in range(ROWS):
		for c in range(COLS):
			var cell: Dictionary = farm[r][c]
			if not _is_cell_unlocked(cell):
				continue
			if cell["crop_id"] != -1 and cell["progress"] < 1.0:
				var gt: float = float(CROPS[int(cell["crop_id"])][3])
				cell["progress"] = minf(cell["progress"] + delta / gt, 1.0)
			cell["wet_timer"] = maxf(float(cell.get("wet_timer", 0.0)) - delta, 0.0)
	save_timer += delta
	if save_timer >= 5.0:
		save_timer = 0.0
		_save_game(false)
	if toast_timer > 0.0:
		toast_timer -= delta
		if toast_timer <= 0.0:
			toast_text = ""
	queue_redraw()

func _notification(what: int):
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		_save_game(false)

func _unhandled_input(event: InputEvent):
	if event is InputEventKey and event.pressed:
		if event.keycode == KEY_S:
			_save_game()
			queue_redraw()
			return

	# --- Mouse button up ---
	if event is InputEventMouseButton and not event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		mouse_held = false
		last_action_col = -1
		last_action_row = -1
		return

	# --- Mouse motion ---
	if event is InputEventMouseMotion:
		var mx: float = event.position.x
		var my: float = event.position.y
		# Hover tracking
		hover_col = -1
		hover_row = -1
		for row in range(ROWS):
			for col in range(COLS):
				var sp := _get_plot_position(col, row)
				if in_diamond(mx, my, sp.x, sp.y):
					hover_col = col
					hover_row = row
					break
			if hover_col >= 0:
				break
		# Drag action: if mouse held and moved to a NEW tile, do action
		if mouse_held and hover_col >= 0 and not shop_open and not inventory_open:
			if hover_col != last_action_col or hover_row != last_action_row:
				_do_tile_action(hover_col, hover_row)
				last_action_col = hover_col
				last_action_row = hover_row
		queue_redraw()
		return

	# --- Mouse button down ---
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		var mx: float = event.position.x
		var my: float = event.position.y
		mouse_held = true

		# Check inventory overlay
		if inventory_open:
			if mx < 200 or mx > 1080 or my < 60 or my > 660:
				inventory_open = false
				queue_redraw()
				return
			if mx >= 915 and mx <= 1075 and my >= 100 and my <= 130:
				_sell_all_inventory()
				queue_redraw()
				return
			# Click on item to sell
			_handle_inventory_click(mx, my)
			return

		if shop_open:
			if mx < 200 or mx > 1080 or my < 60 or my > 660:
				shop_open = false
			else:
				var si := int((my - 150) / 52)
				if my >= 150 and my < 150 + CROPS.size() * 52 and si >= 0 and si < CROPS.size():
					if mx >= 900 and mx <= 960:
						selected_seed = si
						shop_open = false
						toast_text = "已选择: " + str(CROPS[si][0])
						toast_timer = 2.0
					elif mx >= 968 and mx <= 1028:
						_sell_inventory_crop(si, 1)
					elif mx >= 1036 and mx <= 1108:
						_sell_inventory_crop(si, int(inventory.get(si, 0)))
			queue_redraw()
			return

		# Check grid tiles
		for row in range(ROWS):
			for col in range(COLS):
				var sp := _get_plot_position(col, row)
				if in_diamond(mx, my, sp.x, sp.y):
					_do_tile_action(col, row)
					last_action_col = col
					last_action_row = row
					queue_redraw()
					return

		# Panel buttons (single click only, no drag)
		mouse_held = false
		var PX := 890.0
		if side_panel_open:
			if mx >= PX + 450 and mx <= PX + 510 and my >= 50 and my <= 76:
				side_panel_open = false
				shop_open = false
				inventory_open = false
				queue_redraw()
				return
		else:
			if mx >= VIEW_W - 86 and mx <= VIEW_W - 18 and my >= 48 and my <= 82:
				side_panel_open = true
				queue_redraw()
				return
			selected_seed = -1
			queue_redraw()
			return
		if mx >= PX and mx <= PX + 185 and my >= 118 and my <= 146:
			shop_open = true
			queue_redraw()
			return
		# Check buttons explicitly based on the drawn rects
		# 一键收获 button: Rect2(PX + 200, 118, 185, 28)
		if mx >= PX + 200 and mx <= PX + 385 and my >= 118 and my <= 146:
			_harvest_all()
			queue_redraw()
			return
		# 背包 button: Rect2(PX + 400, 118, 110, 28)
		if mx >= PX + 400 and mx <= PX + 510 and my >= 118 and my <= 146:
			inventory_open = true
			shop_open = false
			queue_redraw()
			return
		# Tool mode buttons (below grid)
		var tb_total2: float = 5.0 * 70.0 + 4.0 * 12.0
		var tb_sx: float = OX - tb_total2 * 0.5 + TW * 0.5
		var tb_y2: float = OY + (COLS + ROWS) * TH * 0.5 + 30
		if my >= tb_y2 and my <= tb_y2 + 88:
			for ti in range(5):
				var bx2: float = tb_sx + ti * 82.0
				if mx >= bx2 and mx <= bx2 + 70:
					tool_mode = ti
					var mode_names2 := ["普通", "浇水", "施肥", "收获", "铲除"]
					toast_text = "切换到: " + mode_names2[ti] + "模式"
					toast_timer = 1.0
					queue_redraw()
					return
		# Seed list
		if mx >= PX - 5 and mx <= PX + 515 and my >= 168:
			var si := int((my - 168) / 36)
			if my < 168 + CROPS.size() * 36 and si >= 0 and si < CROPS.size():
				selected_seed = si
				toast_text = "已选择: " + str(CROPS[si][0])
				toast_timer = 2.0
				queue_redraw()
				return
		selected_seed = -1
		queue_redraw()

func _do_tile_action(col: int, row: int):
	var cell: Dictionary = farm[row][col]
	if not _is_cell_unlocked(cell):
		_try_reclaim_plot(col, row)
		return
	# 空地 → 始终可以种植（任何模式下）
	if cell["crop_id"] == -1:
		if tool_mode == 4:
			toast_text = "这里没有作物可以铲除"
			toast_timer = 1.2
			return
		if selected_seed >= 0 and gold >= int(CROPS[selected_seed][1]):
			gold -= int(CROPS[selected_seed][1])
			cell["crop_id"] = selected_seed
			cell["progress"] = 0.0
			cell["wet_timer"] = 0.0
			toast_text = "种下了 " + str(CROPS[selected_seed][0]) + "!"
			toast_timer = 1.5
		else:
			toast_text = "未选择种子或金币不足!"
			toast_timer = 1.5
		return

	var cid: int = int(cell["crop_id"])
	var prog: float = cell["progress"]

	# 普通模式: 不对作物做任何事
	if tool_mode == 0:
		toast_text = "当前是普通模式，切换工具操作作物"
		toast_timer = 1.0
		return

	# 浇水模式
	elif tool_mode == 1:
		if prog >= 1.0:
			toast_text = "作物已成熟，请切换到收获模式!"
			toast_timer = 1.5
		else:
			cell["progress"] = minf(prog + 0.1, 1.0)
			cell["wet_timer"] = 12.0
			toast_text = "浇水成功! 生长+10%"
			toast_timer = 1.5

	# 施肥模式
	elif tool_mode == 2:
		if prog >= 1.0:
			toast_text = "作物已成熟，请切换到收获模式!"
			toast_timer = 1.5
		else:
			var cost: int = int(CROPS[cid][1])
			if gold >= cost:
				gold -= cost
				cell["progress"] = minf(prog + 0.25, 1.0)
				cell["wet_timer"] = 16.0
				toast_text = "施肥成功! 生长+25% (-" + str(cost) + "金币)"
				toast_timer = 1.5
			else:
				toast_text = "金币不足! 施肥需要" + str(cost) + "金币"
				toast_timer = 1.5

	# 收获模式
	elif tool_mode == 3:
		if prog >= 1.0:
			exp_val += int(CROPS[cid][2]) / 5
			_add_to_inventory(cid, 1)
			_add_land_work(row, col, 1)
			_check_lv()
			toast_timer = 1.5
			cell["crop_id"] = -1
			cell["progress"] = 0.0
			cell["wet_timer"] = 0.0
		else:
			toast_text = "作物还没成熟，还需要等待!"
			toast_timer = 1.5

	# 铲除模式
	elif tool_mode == 4:
		toast_text = "铲除了 " + str(CROPS[cid][0])
		toast_timer = 1.5
		cell["crop_id"] = -1
		cell["progress"] = 0.0
		cell["wet_timer"] = 0.0

func _try_reclaim_plot(col: int, row: int):
	var required_level := _get_reclaim_level(col, row)
	var cost := _get_reclaim_cost(col, row)
	if level < required_level:
		toast_text = "等级不足! 这块地需要等级 " + str(required_level)
		toast_timer = 1.8
		return
	if gold < cost:
		toast_text = "金币不足! 开垦需要 " + str(cost) + " 金币"
		toast_timer = 1.8
		return
	gold -= cost
	farm[row][col]["unlocked"] = true
	farm[row][col]["land_level"] = 1
	farm[row][col]["land_work"] = 0
	farm[row][col]["crop_id"] = -1
	farm[row][col]["progress"] = 0.0
	farm[row][col]["wet_timer"] = 0.0
	toast_text = "开垦成功! 解锁第 " + str(_get_plot_index(col, row) + 1) + " 块地"
	toast_timer = 2.0
	_save_game(false)

func _add_land_work(row: int, col: int, amount: int):
	var cell: Dictionary = farm[row][col]
	var land_level := int(cell.get("land_level", LAND_LEVEL_LOCKED))
	if land_level <= LAND_LEVEL_LOCKED or land_level >= LAND_LEVEL_MAX:
		return
	cell["land_work"] = int(cell.get("land_work", 0)) + amount
	while int(cell["land_work"]) >= LAND_UPGRADE_WORK_REQUIRED and int(cell["land_level"]) < LAND_LEVEL_MAX:
		cell["land_work"] = int(cell["land_work"]) - LAND_UPGRADE_WORK_REQUIRED
		cell["land_level"] = int(cell["land_level"]) + 1
		toast_text = "土地升级到 " + _get_land_level_name(int(cell["land_level"])) + "!"
		toast_timer = 2.0
	if int(cell["land_level"]) >= LAND_LEVEL_MAX:
		cell["land_work"] = 0

func _harvest_all():
	var count := 0
	var total := 0
	for r in range(ROWS):
		for c in range(COLS):
			var cell: Dictionary = farm[r][c]
			if not _is_cell_unlocked(cell):
				continue
			if cell["crop_id"] != -1 and cell["progress"] >= 1.0:
				var cid: int = int(cell["crop_id"])
				exp_val += int(CROPS[cid][2]) / 5
				_add_to_inventory(cid, 1)
				_add_land_work(r, c, 1)
				cell["crop_id"] = -1
				cell["progress"] = 0.0
				cell["wet_timer"] = 0.0
				count += 1
	if count > 0:
		_check_lv()
		toast_text = "一键收获了 " + str(count) + " 个作物，已放入背包"
		toast_timer = 2.0
	else:
		toast_text = "没有成熟的作物!"
		toast_timer = 2.0


func _add_to_inventory(cid: int, amount: int):
	var key_str = str(cid)
	if not inventory.has(cid):
		inventory[cid] = 0
	inventory[cid] += amount
	toast_text = "获得 " + str(CROPS[cid][0]) + " x" + str(amount) + "，已放入背包"
	toast_timer = 1.5

func _sell_inventory_crop(cid: int, amount: int) -> int:
	var have := int(inventory.get(cid, 0))
	var sell_amount := mini(maxi(amount, 0), have)
	if sell_amount <= 0:
		toast_text = "背包里没有 " + str(CROPS[cid][0])
		toast_timer = 1.5
		return 0
	gold += int(CROPS[cid][2]) * sell_amount
	inventory[cid] = have - sell_amount
	toast_text = "售出 " + str(CROPS[cid][0]) + " x" + str(sell_amount) + "，获得 " + str(int(CROPS[cid][2]) * sell_amount) + " 金币"
	toast_timer = 1.5
	return sell_amount

func _sell_all_inventory():
	var total_count := 0
	var total_gold := 0
	for cid in inventory.keys():
		var count := int(inventory[cid])
		if count <= 0:
			continue
		total_count += count
		total_gold += int(CROPS[int(cid)][2]) * count
		inventory[cid] = 0
	if total_count <= 0:
		toast_text = "背包是空的"
	else:
		gold += total_gold
		toast_text = "全部卖出 " + str(total_count) + " 个作物，获得 " + str(total_gold) + " 金币"
	toast_timer = 1.8

func _handle_inventory_click(mx: float, my: float):
	var keys = inventory.keys()
	var item_count = 0
	for cid in keys:
			if inventory.has(cid) and inventory[cid] > 0:
				var col = item_count % 5
				var row = item_count / 5
				var slot_x = 200 + col * 150
				var slot_y = 120 + row * 190
				
				# Sell button rect
				if mx >= slot_x + 18 and mx <= slot_x + 67 and my >= slot_y + 123 and my <= slot_y + 148:
					_sell_inventory_crop(int(cid), 1)
					queue_redraw()
					return
				if mx >= slot_x + 73 and mx <= slot_x + 122 and my >= slot_y + 123 and my <= slot_y + 148:
					_sell_inventory_crop(int(cid), int(inventory.get(cid, 0)))
					queue_redraw()
					return
			item_count += 1

func _check_lv():
	while exp_val >= exp_to_level:
		exp_val -= exp_to_level
		level += 1
		exp_to_level = int(exp_to_level * 1.5)

func _get_unlocked_plot_count() -> int:
	var count := 0
	for r in range(ROWS):
		for c in range(COLS):
			if _is_cell_unlocked(farm[r][c]):
				count += 1
	return count

func _get_next_locked_plot() -> Vector2i:
	for r in range(ROWS):
		for c in range(COLS):
			if not _is_cell_unlocked(farm[r][c]):
				return Vector2i(c, r)
	return Vector2i(-1, -1)

func _save_game(show_toast := true):
	var data := {
		"gold": gold,
		"level": level,
		"exp_val": exp_val,
		"exp_to_level": exp_to_level,
		"farm": farm,
		"inventory": inventory,
		"selected_seed": selected_seed,
		"tool_mode": tool_mode,
		"saved_at": Time.get_unix_time_from_system(),
	}
	var file := FileAccess.open(SAVE_PATH, FileAccess.WRITE)
	if file == null:
		if show_toast:
			toast_text = "保存失败"
			toast_timer = 1.5
		return
	file.store_string(JSON.stringify(data))
	if show_toast:
		toast_text = "农场已保存"
		toast_timer = 1.2

func _load_game():
	if not FileAccess.file_exists(SAVE_PATH):
		return
	var file := FileAccess.open(SAVE_PATH, FileAccess.READ)
	if file == null:
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if not (parsed is Dictionary):
		return
	var data: Dictionary = parsed
	gold = int(data.get("gold", gold))
	level = int(data.get("level", level))
	exp_val = int(data.get("exp_val", exp_val))
	exp_to_level = int(data.get("exp_to_level", exp_to_level))
	selected_seed = int(data.get("selected_seed", selected_seed))
	tool_mode = int(data.get("tool_mode", tool_mode))
	if data.has("inventory") and (data["inventory"] is Dictionary):
		inventory = data["inventory"]
		_normalize_inventory_keys()
	if data.has("farm") and (data["farm"] is Array):
		_restore_farm(data["farm"])
	_apply_offline_growth(int(data.get("saved_at", Time.get_unix_time_from_system())))

func _normalize_inventory_keys():
	var fixed := {}
	for key in inventory.keys():
		var cid := int(key)
		fixed[cid] = int(inventory[key])
	inventory = fixed

func _restore_farm(saved_farm: Array):
	for r in range(mini(ROWS, saved_farm.size())):
		if not (saved_farm[r] is Array):
			continue
		var saved_row: Array = saved_farm[r]
		for c in range(mini(COLS, saved_row.size())):
			if saved_row[c] is Dictionary:
				var saved_cell: Dictionary = saved_row[c]
				var cid := int(saved_cell.get("crop_id", -1))
				if cid >= -1 and cid < CROPS.size():
					farm[r][c]["crop_id"] = cid
					farm[r][c]["progress"] = clampf(float(saved_cell.get("progress", 0.0)), 0.0, 1.0)
					farm[r][c]["wet_timer"] = maxf(float(saved_cell.get("wet_timer", 0.0)), 0.0)
					var was_unlocked := bool(saved_cell.get("unlocked", cid != -1 or _get_plot_index(c, r) < INITIAL_UNLOCKED_PLOTS))
					var land_level := int(saved_cell.get("land_level", 1 if was_unlocked else LAND_LEVEL_LOCKED))
					land_level = clampi(land_level, LAND_LEVEL_LOCKED, LAND_LEVEL_MAX)
					farm[r][c]["land_level"] = land_level
					farm[r][c]["land_work"] = clampi(int(saved_cell.get("land_work", 0)), 0, LAND_UPGRADE_WORK_REQUIRED - 1)
					farm[r][c]["unlocked"] = land_level > LAND_LEVEL_LOCKED

func _apply_offline_growth(saved_at: int):
	var elapsed: int = maxi(0, int(Time.get_unix_time_from_system()) - saved_at)
	if elapsed <= 0:
		return
	var grew_count := 0
	for r in range(ROWS):
		for c in range(COLS):
			var cell: Dictionary = farm[r][c]
			if not _is_cell_unlocked(cell):
				continue
			if cell["crop_id"] != -1 and cell["progress"] < 1.0:
				var cid := int(cell["crop_id"])
				var before := float(cell["progress"])
				var gt := float(CROPS[cid][3])
				cell["progress"] = minf(before + float(elapsed) / gt, 1.0)
				if before < 1.0 and cell["progress"] >= 1.0:
					grew_count += 1
	if grew_count > 0:
		toast_text = "离线期间成熟了 " + str(grew_count) + " 个作物"
		toast_timer = 3.0

# ===================== DRAW =====================
func _draw():
	# Title bar
	draw_rect(Rect2(100, 40, 440, 40), Color(0.1, 0.06, 0.02, 0.85))
	_draw_text(200, 46, "QQ 农场 2.5D", 24, Color(1, 0.9, 0.2))

	# ---- ISOMETRIC TILES: PASS 1 - Soil + borders (back to front) ----
	for row in range(ROWS):
		for col in range(COLS):
			var sp := _get_plot_position(col, row)
			var cx: float = sp.x
			var cy: float = sp.y
			var corners := iso_corners(cx, cy)
			var cell: Dictionary = farm[row][col]
			_draw_land_tile(corners, cell)

			# Hover glow (on soil layer)
			if col == hover_col and row == hover_row:
				if not _is_cell_unlocked(cell):
					draw_colored_polygon(corners, Color(0.75, 0.75, 0.75, 0.22))
				elif cell["crop_id"] == -1 and selected_seed >= 0:
					draw_colored_polygon(corners, Color(0.3, 0.9, 0.3, 0.25))
				elif cell["crop_id"] != -1 and cell["progress"] >= 1.0:
					draw_colored_polygon(corners, Color(1.0, 0.85, 0.15, 0.35))
				else:
					draw_colored_polygon(corners, Color(1, 1, 1, 0.1))

			# Border
			if col == hover_col and row == hover_row:
				var bcol := Color(1, 0.9, 0.2, 0.9)
				for i in range(4):
					draw_line(corners[i], corners[(i + 1) % 4], bcol, 2.0)

			# Seed preview on empty tile
			if _is_cell_unlocked(cell) and cell["crop_id"] == -1 and col == hover_col and row == hover_row and selected_seed >= 0:
				var seed_texture := _get_crop_seed_texture(selected_seed)
				if seed_texture != null:
					_draw_seed_preview_texture(cx, cy, seed_texture)
				else:
					draw_circle(Vector2(cx, cy), 10, Color(CROP_COLORS[selected_seed][1].r, CROP_COLORS[selected_seed][1].g, CROP_COLORS[selected_seed][1].b, 0.7))
					draw_circle(Vector2(cx, cy), 6, CROP_COLORS[selected_seed][1])

	# ---- ISOMETRIC TILES: PASS 2 - Crops + progress bars (back to front) ----
	for row in range(ROWS):
		for col in range(COLS):
			var sp := _get_plot_position(col, row)
			var cx: float = sp.x
			var cy: float = sp.y
			var cell: Dictionary = farm[row][col]

			if not _is_cell_unlocked(cell):
				var req_level := _get_reclaim_level(col, row)
				var req_cost := _get_reclaim_cost(col, row)
				var lock_text := "Lv" + str(req_level)
				var f_lock: Font = ThemeDB.fallback_font
				var lock_w: float = f_lock.get_string_size(lock_text, HORIZONTAL_ALIGNMENT_LEFT, -1, 12).x
				draw_rect(Rect2(cx - lock_w * 0.5 - 6, cy - 10, lock_w + 12, 18), Color(0.02, 0.02, 0.02, 0.68))
				_draw_text(cx - lock_w * 0.5, cy - 8, lock_text, 12, Color(0.92, 0.9, 0.78, 0.95))
				if col == hover_col and row == hover_row:
					var cost_text := str(req_cost) + " 金币开垦"
					var cost_w: float = f_lock.get_string_size(cost_text, HORIZONTAL_ALIGNMENT_LEFT, -1, 11).x
					draw_rect(Rect2(cx - cost_w * 0.5 - 6, cy + 12, cost_w + 12, 17), Color(0.02, 0.02, 0.02, 0.72))
					_draw_text(cx - cost_w * 0.5, cy + 13, cost_text, 11, Color(1.0, 0.82, 0.26, 0.95))
				continue

			if cell["crop_id"] != -1:
				var cid: int = int(cell["crop_id"])
				var prog: float = cell["progress"]
				var fruit_col: Color = CROP_COLORS[cid][1]
				var leaf_col: Color = CROP_COLORS[cid][0]
				var stage: int = _get_growth_stage(prog)
				var atlas_texture: Texture2D = _get_crop_stage_texture(cid, prog)

				if stage < 0:
					_draw_plant_seed(cx, cy, prog)
				elif atlas_texture != null:
					_draw_crop_atlas_texture(cx, cy, atlas_texture, prog, cid, stage)
				elif prog > 0.3:
					_draw_plant_growing(cx, cy, leaf_col, prog)
				else:
					_draw_plant_seed(cx, cy, prog)

				if prog >= 1.0:
					if atlas_texture == null:
						_draw_plant_full(cx, cy, leaf_col, fruit_col)
					# Harvest label
					var f: Font = ThemeDB.fallback_font
					var lbl := "收获"
					var lw: float = f.get_string_size(lbl, HORIZONTAL_ALIGNMENT_LEFT, -1, 11).x
					draw_rect(Rect2(cx - lw * 0.5 - 4, cy - TH * 0.5 - 22, lw + 8, 16), Color(0.8, 0.6, 0, 0.85))
					_draw_text(cx - lw * 0.5, cy - TH * 0.5 - 19, lbl, 11, Color(1, 1, 1))

				# Progress bar
				if prog < 1.0:
					var bw: float = TW * 0.5
					var bx: float = cx - bw * 0.5
					var by: float = cy + TH * 0.35
					draw_rect(Rect2(bx, by, bw, 5), Color(0, 0, 0, 0.5))
					var bc := Color(0.2, 0.8, 0.3) if prog < 0.6 else Color(0.95, 0.75, 0.1)
					draw_rect(Rect2(bx, by, bw * prog, 5), bc)
					# Stage name + time remaining (always visible)
					var stage_name: String
					var remaining: int = int((1.0 - prog) * float(CROPS[cid][3]))
					if prog < 0.15:
						stage_name = "种子"
					elif prog < 0.4:
						stage_name = "发芽"
					elif prog < 0.7:
						stage_name = "生长"
					else:
						stage_name = "快熟"
					var info: String = stage_name + " " + str(remaining) + "秒"
					_draw_text(cx - 20, cy + TH * 0.5 + 14, info, 9, Color(1, 1, 1, 0.75))
				else:
					# Mature: show price hint
					_draw_text(cx - 18, cy + TH * 0.5 + 14, "售价 " + str(int(CROPS[cid][2])), 9, Color(1, 0.9, 0.3, 0.8))

	# ---- HOVER TOOLTIP (QQ Farm style detail card) ----
	if hover_col >= 0 and hover_col < COLS and hover_row >= 0 and hover_row < ROWS:
		var hcell: Dictionary = farm[hover_row][hover_col]
		if not _is_cell_unlocked(hcell):
			var hsp_locked := _get_plot_position(hover_col, hover_row)
			var lx: float = hsp_locked.x
			var ly: float = hsp_locked.y
			var ltw: float = 190.0
			var lth: float = 82.0
			var ltx: float = clampf(lx - ltw * 0.5, 5, VIEW_W - ltw - 5)
			var lty: float = clampf(ly - TH * 0.5 - lth - 18, 5, VIEW_H - lth - 5)
			var required_level := _get_reclaim_level(hover_col, hover_row)
			var required_cost := _get_reclaim_cost(hover_col, hover_row)
			draw_rect(Rect2(ltx, lty, ltw, lth), Color(0.08, 0.06, 0.03, 0.94))
			draw_rect(Rect2(ltx, lty, ltw, lth), Color(0.55, 0.42, 0.2), false, 2)
			draw_rect(Rect2(ltx, lty, ltw, 22), Color(0.34, 0.25, 0.12))
			_draw_text(ltx + 8, lty + 3, "未开垦土地", 13, Color(1.0, 0.92, 0.72))
			_draw_text(ltx + 10, lty + 31, "需要等级: " + str(required_level), 12, Color(0.86, 0.92, 1.0))
			_draw_text(ltx + 10, lty + 49, "开垦费用: " + str(required_cost) + " 金币", 12, Color(1.0, 0.84, 0.25))
			var locked_hint := "点击开垦" if level >= required_level and gold >= required_cost else "等级或金币不足"
			var locked_hint_color := Color(0.45, 0.95, 0.45) if level >= required_level and gold >= required_cost else Color(0.95, 0.45, 0.35)
			_draw_text(ltx + 10, lty + 66, locked_hint, 10, locked_hint_color)
		elif hcell["crop_id"] != -1:
			var hsp := _get_plot_position(hover_col, hover_row)
			var hx: float = hsp.x
			var hy: float = hsp.y
			var hid: int = int(hcell["crop_id"])
			var hprog: float = hcell["progress"]

			# Stage info
			var stage: String
			var stage_color: Color
			if hprog >= 1.0:
				stage = "可以收获"
				stage_color = Color(1, 0.85, 0.1)
			elif hprog >= 0.7:
				stage = "即将成熟"
				stage_color = Color(0.9, 0.8, 0.2)
			elif hprog >= 0.4:
				stage = "生长中"
				stage_color = Color(0.3, 0.8, 0.3)
			elif hprog >= 0.15:
				stage = "发芽中"
				stage_color = Color(0.4, 0.75, 0.3)
			else:
				stage = "刚种下"
				stage_color = Color(0.6, 0.6, 0.6)

			var time_left: int = int((1.0 - hprog) * float(CROPS[hid][3]))
			var pct: int = int(hprog * 100)

			# Tooltip position (above the tile, clamp to screen)
			var tw: float = 170.0
			var th: float = 95.0
			var tx: float = hx - tw * 0.5
			var ty: float = hy - TH * 0.5 - th - 18
			# Clamp to screen
			tx = clampf(tx, 5, VIEW_W - tw - 5)
			ty = clampf(ty, 5, VIEW_H - th - 5)

			# Card background
			draw_rect(Rect2(tx, ty, tw, th), Color(0.08, 0.05, 0.02, 0.92))
			draw_rect(Rect2(tx, ty, tw, th), Color(0.55, 0.42, 0.2), false, 2)

			# Title bar
			draw_rect(Rect2(tx, ty, tw, 22), Color(0.4, 0.28, 0.1))
			_draw_text(tx + 6, ty + 3, str(CROPS[hid][0]), 13, Color(1, 0.95, 0.8))

			# Crop color dot
			draw_circle(Vector2(tx + 14, ty + 36), 6, CROP_COLORS[hid][1])

			# Stage
			_draw_text(tx + 26, ty + 30, stage, 11, stage_color)

			# Progress percentage
			_draw_text(tx + 26, ty + 46, str(pct) + "% 已成长", 11, Color(0.8, 0.8, 0.8))
			var land_info := _get_land_level_name(int(hcell.get("land_level", 1)))
			if int(hcell.get("land_level", 1)) < LAND_LEVEL_MAX:
				land_info += " " + str(int(hcell.get("land_work", 0))) + "/" + str(LAND_UPGRADE_WORK_REQUIRED)
			_draw_text(tx + 92, ty + 30, land_info, 10, Color(0.95, 0.82, 0.42))

			# Time / Price + action hint
			var action_hint: String
			if hprog >= 1.0:
				_draw_text(tx + 26, ty + 62, "价值: " + str(int(CROPS[hid][2])) + " 金币", 11, Color(1, 0.88, 0.15))
				if tool_mode == 3:
					action_hint = "点击收获!"
				elif tool_mode == 4:
					action_hint = "点击铲除作物"
				else:
					action_hint = "切换到收获模式"
			else:
				_draw_text(tx + 26, ty + 62, "剩余时间: " + str(time_left) + "秒", 11, Color(0.7, 0.8, 1.0))
				if tool_mode == 1:
					action_hint = "点击浇水 (+10%)"
				elif tool_mode == 2:
					action_hint = "点击施肥 (+25%) -" + str(int(CROPS[hid][1])) + "金币"
				elif tool_mode == 4:
					action_hint = "点击铲除作物"
				else:
					action_hint = "切换到浇水/施肥模式"
			var hint_color := Color(0.4, 0.9, 0.4) if hprog >= 1.0 else Color(0.4, 0.7, 0.9)
			_draw_text(tx + 26, ty + 76, action_hint, 10, hint_color)

			# Small arrow pointing down to tile
			var arrow_x: float = clampf(hx, tx + 10, tx + tw - 10)
			var arrow_pts: PackedVector2Array = PackedVector2Array([
				Vector2(arrow_x - 6, ty + th),
				Vector2(arrow_x, ty + th + 8),
				Vector2(arrow_x + 6, ty + th),
			])
			draw_colored_polygon(arrow_pts, Color(0.08, 0.05, 0.02, 0.92))
		else:
			_draw_land_tooltip(hover_col, hover_row)

	# ---- TOOL BAR (地块下方图标按钮) ----
	var tb_names := ["普通", "浇水", "施肥", "收获", "铲除"]
	var tb_colors := [
		Color(0.5, 0.47, 0.42),
		Color(0.22, 0.52, 0.88),
		Color(0.78, 0.58, 0.12),
		Color(0.22, 0.72, 0.32),
		Color(0.64, 0.38, 0.22),
	]
	var tb_dark := [
		Color(0.35, 0.32, 0.28),
		Color(0.14, 0.38, 0.68),
		Color(0.58, 0.4, 0.06),
		Color(0.14, 0.52, 0.2),
		Color(0.42, 0.24, 0.14),
	]
	var btn_size: float = 70.0
	var btn_gap: float = 12.0
	var tb_total: float = 5.0 * btn_size + 4.0 * btn_gap
	var tb_start_x: float = OX - tb_total * 0.5 + TW * 0.5
	var tb_y: float = OY + (COLS + ROWS) * TH * 0.5 + 30

	for ti in range(5):
		var bx: float = tb_start_x + ti * (btn_size + btn_gap)
		var by: float = tb_y

		var is_active: bool = (tool_mode == ti)

		# ---- Draw icon in each button ----
		var icon_cx: float = bx + btn_size * 0.5
		var icon_cy: float = by + btn_size * 0.35
		var icon_texture: Texture2D = TOOL_ICON_TEXTURES[ti]
		if icon_texture != null:
			var icon_grow := 4.0 if is_active else 0.0
			var icon_rect := Rect2(bx - icon_grow * 0.5, by - icon_grow * 0.5, btn_size + icon_grow, btn_size + icon_grow)
			if is_active:
				draw_circle(Vector2(icon_cx, by + btn_size * 0.48), 38.0, Color(1.0, 0.86, 0.22, 0.18))
				draw_circle(Vector2(icon_cx, by + btn_size * 0.48), 30.0, Color(1.0, 0.96, 0.42, 0.10))
			draw_texture_rect(icon_texture, icon_rect, false)
			var txt_col: Color = Color(1.0, 0.92, 0.34, 0.98) if is_active else Color(0.78, 0.78, 0.78)
			var label_w: float = ThemeDB.fallback_font.get_string_size(tb_names[ti], HORIZONTAL_ALIGNMENT_LEFT, -1, 12).x
			_draw_text(icon_cx - label_w * 0.5, by + btn_size + 1.0, tb_names[ti], 12, txt_col)
			if is_active:
				draw_line(Vector2(icon_cx - 15.0, by + btn_size + 17.0), Vector2(icon_cx + 15.0, by + btn_size + 17.0), Color(1.0, 0.86, 0.22, 0.95), 3.0)
			continue

		if ti == 0:
			# 普通 - 鼠标箭头
			var arrow: PackedVector2Array = PackedVector2Array([
				Vector2(icon_cx - 6, icon_cy - 10),
				Vector2(icon_cx - 6, icon_cy + 8),
				Vector2(icon_cx - 1, icon_cy + 4),
				Vector2(icon_cx + 4, icon_cy + 10),
				Vector2(icon_cx + 7, icon_cy + 8),
				Vector2(icon_cx + 2, icon_cy + 2),
				Vector2(icon_cx + 8, icon_cy - 2),
			])
			draw_colored_polygon(arrow, Color(1, 1, 1, 0.9))

		elif ti == 1:
			# 浇水 - 水壶
			draw_rect(Rect2(icon_cx - 8, icon_cy - 4, 16, 10), Color(0.6, 0.75, 0.9))
			draw_rect(Rect2(icon_cx - 8, icon_cy - 4, 16, 10), Color(0.3, 0.5, 0.8), false, 1.5)
			# 壶嘴
			draw_line(Vector2(icon_cx + 8, icon_cy - 4), Vector2(icon_cx + 14, icon_cy - 10), Color(0.3, 0.5, 0.8), 2.0)
			# 水滴
			draw_circle(Vector2(icon_cx + 10, icon_cy - 14), 2.5, Color(0.3, 0.6, 1.0))
			draw_circle(Vector2(icon_cx + 5, icon_cy - 16), 2.0, Color(0.3, 0.6, 1.0))
			# 把手
			draw_arc(Vector2(icon_cx - 2, icon_cy - 8), 6, 0, PI, 12, Color(0.3, 0.5, 0.8), 2.0)

		elif ti == 2:
			# 施肥 - 袋子
			var bag_body: PackedVector2Array = PackedVector2Array([
				Vector2(icon_cx - 11, icon_cy - 7),
				Vector2(icon_cx + 11, icon_cy - 7),
				Vector2(icon_cx + 9, icon_cy + 12),
				Vector2(icon_cx - 9, icon_cy + 12),
			])
			draw_colored_polygon(bag_body, Color(0.68, 0.52, 0.23))
			for bi in range(4):
				draw_line(bag_body[bi], bag_body[(bi + 1) % 4], Color(0.42, 0.28, 0.08), 1.5)
			draw_line(Vector2(icon_cx - 6, icon_cy - 8), Vector2(icon_cx + 6, icon_cy - 8), Color(0.32, 0.22, 0.08), 2.0)
			_draw_text(icon_cx - 6, icon_cy - 2, "肥", 11, Color(1, 0.92, 0.55))
			draw_circle(Vector2(icon_cx - 15, icon_cy + 12), 2.0, Color(0.9, 0.75, 0.25))
			draw_circle(Vector2(icon_cx + 14, icon_cy + 10), 1.8, Color(0.9, 0.75, 0.25))

		elif ti == 3:
			# 收获 - 篮子
			# 篮身 (梯形)
			var basket: PackedVector2Array = PackedVector2Array([
				Vector2(icon_cx - 10, icon_cy - 4),
				Vector2(icon_cx + 10, icon_cy - 4),
				Vector2(icon_cx + 8, icon_cy + 8),
				Vector2(icon_cx - 8, icon_cy + 8),
			])
			draw_colored_polygon(basket, Color(0.7, 0.55, 0.25))
			# 提手
			draw_arc(Vector2(icon_cx, icon_cy - 10), 8, PI, TAU, 12, Color(0.45, 0.32, 0.1), 2.0)
			# 里面的小果子
			draw_circle(Vector2(icon_cx - 3, icon_cy), 2.5, Color(0.9, 0.2, 0.15))
			draw_circle(Vector2(icon_cx + 3, icon_cy - 1), 2.5, Color(1.0, 0.6, 0.1))
			draw_circle(Vector2(icon_cx, icon_cy - 5), 2.4, Color(0.95, 0.92, 0.2))

		elif ti == 4:
			# 铲除 - 铲子
			draw_line(Vector2(icon_cx + 8, icon_cy - 16), Vector2(icon_cx - 8, icon_cy + 7), Color(0.92, 0.74, 0.42), 4.0)
			draw_line(Vector2(icon_cx + 8, icon_cy - 16), Vector2(icon_cx - 8, icon_cy + 7), Color(0.42, 0.24, 0.12), 1.2)
			var blade: PackedVector2Array = PackedVector2Array([
				Vector2(icon_cx - 14, icon_cy + 7),
				Vector2(icon_cx - 3, icon_cy + 2),
				Vector2(icon_cx + 3, icon_cy + 12),
				Vector2(icon_cx - 5, icon_cy + 19),
			])
			draw_colored_polygon(blade, Color(0.78, 0.84, 0.86))
			for si in range(4):
				draw_line(blade[si], blade[(si + 1) % 4], Color(0.38, 0.42, 0.44), 1.4)
			draw_line(Vector2(icon_cx + 5, icon_cy - 19), Vector2(icon_cx + 16, icon_cy - 13), Color(0.42, 0.24, 0.12), 3.0)

		# Tool name below icon
		var txt_col: Color = Color(1, 1, 1, 0.95) if is_active else Color(0.7, 0.7, 0.7)
		_draw_text(bx + 14, by + btn_size - 6, tb_names[ti], 12, txt_col)

	# ---- SIDE PANEL ----
	var PX := 890.0
	if not side_panel_open:
		_draw_side_panel_toggle()
	else:
		_draw_side_panel(PX)

	# HUD top-left
	draw_rect(Rect2(10, 8, 180, 38), Color(0, 0, 0, 0.6))
	_draw_text(18, 14, "金币:" + str(gold) + "  等级:" + str(level), 16, Color(1, 0.88, 0.15))

	# ---- SHOP OVERLAY ----
	if shop_open:
		draw_rect(Rect2(0, 0, VIEW_W, VIEW_H), Color(0, 0, 0, 0.6))
		draw_rect(Rect2(150, 50, 980, 620), Color(0.94, 0.9, 0.78))
		draw_rect(Rect2(150, 50, 980, 620), Color(0.48, 0.32, 0.12), false, 4)
		draw_rect(Rect2(150, 50, 980, 45), Color(0.42, 0.28, 0.08))
		_draw_text(480, 57, "种子商店", 24, Color(1, 0.9, 0.25))
		_draw_text(900, 62, "点击外部关闭", 13, Color(0.75, 0.68, 0.55))
		# Header
		draw_rect(Rect2(170, 110, 940, 30), Color(0.58, 0.48, 0.28))
		_draw_text(220, 115, "名称", 14, Color(1, 1, 1))
		_draw_text(400, 115, "种子价", 14, Color(1, 1, 1))
		_draw_text(540, 115, "售价", 14, Color(1, 1, 1))
		_draw_text(680, 115, "生长时间", 14, Color(1, 1, 1))
		_draw_text(830, 115, "利润", 14, Color(1, 1, 1))
		_draw_text(918, 115, "操作", 14, Color(1, 1, 1))
		for i in range(CROPS.size()):
			var iy: float = 150.0 + float(i) * 52.0
			var bg := Color(0.88, 0.83, 0.68) if i % 2 == 0 else Color(0.91, 0.86, 0.73)
			draw_rect(Rect2(170, iy, 940, 46), bg)
			var shop_seed_texture := _get_crop_seed_texture(i)
			if shop_seed_texture != null:
				_draw_ui_seed_thumbnail(Rect2(190, iy + 7, 34, 34), shop_seed_texture)
			else:
				draw_circle(Vector2(210, iy + 26), 10, CROP_COLORS[i][1])
			_draw_text(230, iy + 8, str(CROPS[i][0]), 16, Color(0.08, 0.08, 0.08))
			_draw_text(400, iy + 12, str(int(CROPS[i][1])) + " 金币", 14, Color(0.75, 0.25, 0.08))
			_draw_text(540, iy + 12, str(int(CROPS[i][2])) + " 金币", 14, Color(0.08, 0.55, 0.08))
			_draw_text(680, iy + 12, str(int(CROPS[i][3])) + " 秒", 14, Color(0.18, 0.18, 0.5))
			var profit: int = int(CROPS[i][2]) - int(CROPS[i][1])
			_draw_text(830, iy + 12, "+" + str(profit), 15, Color(0, 0.55, 0))
			draw_rect(Rect2(900, iy + 9, 60, 28), Color(0.22, 0.62, 0.28))
			_draw_text(914, iy + 14, "购买", 13, Color(1, 1, 1))
			draw_rect(Rect2(968, iy + 9, 60, 28), Color(0.76, 0.50, 0.12))
			_draw_text(982, iy + 14, "卖出", 13, Color(1, 1, 1))
			draw_rect(Rect2(1036, iy + 9, 72, 28), Color(0.62, 0.24, 0.18))
			_draw_text(1044, iy + 14, "全卖", 13, Color(1, 1, 1))
		_draw_text(430, 630, "购买选择种子，卖出使用背包库存", 16, Color(0.38, 0.3, 0.2))

	# ---- INVENTORY OVERLAY ----
	if inventory_open:
		draw_rect(Rect2(0, 0, VIEW_W, VIEW_H), Color(0, 0, 0, 0.6))
		draw_rect(Rect2(150, 50, 980, 620), Color(0.9, 0.86, 0.94))
		draw_rect(Rect2(150, 50, 980, 620), Color(0.4, 0.18, 0.45), false, 4)
		draw_rect(Rect2(150, 50, 980, 45), Color(0.35, 0.15, 0.4))
		_draw_text(480, 57, "我的背包 (售出换金币)", 24, Color(1, 0.9, 0.95))
		_draw_text(900, 62, "点击外部关闭", 13, Color(0.85, 0.75, 0.9))
		draw_rect(Rect2(915, 100, 160, 30), Color(0.65, 0.28, 0.18))
		_draw_text(952, 105, "全部卖出", 15, Color(1, 1, 1))
		
		var keys = inventory.keys()
		var is_empty = true
		var item_count = 0
		for cid in keys:
			if inventory[cid] > 0:
				is_empty = false
				var col = item_count % 5
				var row = item_count / 5
				var slot_x = 200 + col * 150
				var slot_y = 120 + row * 180
				
				# Slot background
				draw_rect(Rect2(slot_x, slot_y, 140, 160), Color(0.85, 0.8, 0.9))
				draw_rect(Rect2(slot_x, slot_y, 140, 160), Color(0.5, 0.3, 0.55), false, 2)
				
				# Crop icon
				draw_circle(Vector2(slot_x + 70, slot_y + 40), 20, CROP_COLORS[cid][1])
				draw_circle(Vector2(slot_x + 63, slot_y + 35), 6, Color(1,1,1,0.3))
				
				# Name and amount
				var name_w = ThemeDB.fallback_font.get_string_size(CROPS[cid][0], HORIZONTAL_ALIGNMENT_LEFT, -1, 16).x
				_draw_text(slot_x + 70 - name_w * 0.5, slot_y + 70, str(CROPS[cid][0]), 16, Color(0.1, 0.1, 0.2))
				var amt_txt = "数量: " + str(inventory[cid])
				var amt_w = ThemeDB.fallback_font.get_string_size(amt_txt, HORIZONTAL_ALIGNMENT_LEFT, -1, 14).x
				_draw_text(slot_x + 70 - amt_w * 0.5, slot_y + 95, amt_txt, 14, Color(0.3, 0.2, 0.4))
				
				# Sell button
				draw_rect(Rect2(slot_x + 35, slot_y + 120, 70, 25), Color(0.8, 0.6, 0.1))
				_draw_text(slot_x + 48, slot_y + 124, "卖出", 14, Color(1,1,1))
				var price_txt = "(+" + str(int(CROPS[cid][2])) + ")"
				var price_w = ThemeDB.fallback_font.get_string_size(price_txt, HORIZONTAL_ALIGNMENT_LEFT, -1, 10).x
				_draw_text(slot_x + 70 - price_w * 0.5, slot_y + 148, price_txt, 10, Color(0.5, 0.4, 0.1))
				
				item_count += 1
				
		if is_empty:
			_draw_text(540, 300, "背包是空的", 20, Color(0.5, 0.4, 0.6))

	# Toast
	if toast_text != "":
		var alpha := minf(toast_timer, 1.0)
		var f: Font = ThemeDB.fallback_font
		var tw: float = f.get_string_size(toast_text, HORIZONTAL_ALIGNMENT_LEFT, -1, 20).x + 50
		var tx: float = (VIEW_W - tw) * 0.5
		draw_rect(Rect2(tx, 670, tw, 38), Color(0, 0, 0, 0.8 * alpha))
		_draw_text(tx + 25, 677, toast_text, 18, Color(1, 1, 1, alpha))

func _draw_side_panel(PX: float):
	# Background
	draw_rect(Rect2(PX - 10, 40, 540, 630), Color(0.10, 0.07, 0.03, 0.9))
	draw_rect(Rect2(PX - 12, 38, 544, 634), Color(0.5, 0.38, 0.18), false, 3)

	# Gold
	_draw_text(PX, 52, "金币: " + str(gold), 26, Color(1, 0.88, 0.15))
	# Level
	_draw_text(PX + 280, 56, "等级 " + str(level), 20, Color(0.85, 0.92, 1.0))
	draw_rect(Rect2(PX + 450, 50, 60, 26), Color(0.22, 0.18, 0.1, 0.95))
	draw_rect(Rect2(PX + 450, 50, 60, 26), Color(0.72, 0.55, 0.24), false, 1.5)
	_draw_text(PX + 464, 54, "收起", 13, Color(1, 0.92, 0.72))
	# Exp bar
	draw_rect(Rect2(PX, 85, 510, 12), Color(0.08, 0.08, 0.12))
	var ep: float = float(exp_val) / float(exp_to_level)
	draw_rect(Rect2(PX, 85, 510.0 * ep, 12), Color(0.28, 0.55, 1.0))
	_draw_text(PX + 380, 84, str(exp_val) + "/" + str(exp_to_level), 11, Color(0.7, 0.8, 1))

	var unlocked_count := _get_unlocked_plot_count()
	var next_plot := _get_next_locked_plot()
	var reclaim_text := "土地: " + str(unlocked_count) + "/" + str(COLS * ROWS)
	if next_plot.x >= 0:
		reclaim_text += "  下一块 Lv" + str(_get_reclaim_level(next_plot.x, next_plot.y)) + " / " + str(_get_reclaim_cost(next_plot.x, next_plot.y)) + "金币"
	else:
		reclaim_text += "  已全部开垦"
	_draw_text(PX, 102, reclaim_text, 12, Color(0.82, 0.78, 0.62))

	# Shop button
	draw_rect(Rect2(PX, 118, 185, 28), Color(0.88, 0.58, 0.1))
	_draw_text(PX + 40, 122, "种子商店", 16, Color(1, 1, 1))
	# Harvest all button
	draw_rect(Rect2(PX + 200, 118, 185, 28), Color(0.18, 0.72, 0.28))
	_draw_text(PX + 230, 122, "一键收获", 16, Color(1, 1, 1))
	# Inventory button
	draw_rect(Rect2(PX + 400, 118, 110, 28), Color(0.68, 0.28, 0.68))
	_draw_text(PX + 430, 122, "打开背包", 16, Color(1, 1, 1))

	# Seed list header
	_draw_text(PX, 148, "选择种子:", 16, Color(0.9, 0.85, 0.7))

	for i in range(CROPS.size()):
		var sy: float = 168.0 + float(i) * 36.0
		# Row background
		if selected_seed == i:
			draw_rect(Rect2(PX - 5, sy, 520, 32), Color(1, 1, 0.3, 0.2))
			draw_rect(Rect2(PX - 5, sy, 520, 32), Color(1, 0.9, 0.2), false, 2)
		else:
			draw_rect(Rect2(PX - 5, sy, 520, 32), Color(0.12, 0.09, 0.05, 0.5))
		var seed_list_texture := _get_crop_seed_texture(i)
		if seed_list_texture != null:
			_draw_ui_seed_thumbnail(Rect2(PX + 4, sy + 2, 24, 24), seed_list_texture)
		else:
			draw_circle(Vector2(PX + 15, sy + 16), 7, CROP_COLORS[i][1])
		# Name
		_draw_text(PX + 30, sy + 2, str(CROPS[i][0]), 14, Color(1, 1, 1))
		# Info
		_draw_text(PX + 150, sy + 5, str(int(CROPS[i][1])) + "/" + str(int(CROPS[i][2])) + "/" + str(int(CROPS[i][3])) + "秒", 11, Color(0.6, 0.6, 0.6))

	# Instructions
	_draw_text(PX, 500, "操作提示:", 14, Color(0.55, 0.5, 0.4))
	_draw_text(PX, 520, "1. 选择种子→点击空地种植", 12, Color(0.5, 0.45, 0.38))
	_draw_text(PX, 538, "2. 切换浇水/施肥/收获/铲除工具", 12, Color(0.5, 0.45, 0.38))
	var preview_index := maxi(selected_seed, 0)
	var preview_crop = CROPS[preview_index]
	_draw_text(PX, 548, "当前种子预览:", 14, Color(0.75, 0.68, 0.55))
	_draw_crop_preview(PX + 10, 570, _get_crop_seed_texture(preview_index), str(preview_crop[0]))
	_draw_text(PX + 150, 572, "已接入 9 种植物", 15, Color(0.95, 0.88, 0.62))
	_draw_text(PX + 150, 594, "排序方式: 按价值从低到高", 12, Color(0.7, 0.66, 0.54))
	_draw_text(PX + 150, 614, "当前选择: " + str(preview_crop[0]), 12, Color(0.85, 0.82, 0.76))

func _draw_side_panel_toggle():
	var rect := Rect2(VIEW_W - 86, 48, 68, 34)
	draw_rect(rect, Color(0.10, 0.07, 0.03, 0.88))
	draw_rect(rect, Color(0.72, 0.55, 0.24), false, 2.0)
	_draw_text(rect.position.x + 14, rect.position.y + 8, "展开", 14, Color(1, 0.92, 0.72))

# ---- PLANT DRAWING ----
func _draw_plant_full(cx: float, cy: float, leaf: Color, fruit: Color):
	var by: float = cy
	var sh: float = 28.0
	# Stem
	draw_line(Vector2(cx, by), Vector2(cx, by - sh), Color(0.2, 0.5, 0.12), 3.0)
	# Leaves
	draw_circle(Vector2(cx - 10, by - sh * 0.6), 8, leaf)
	draw_circle(Vector2(cx + 10, by - sh * 0.55), 7, leaf)
	# Fruit
	draw_circle(Vector2(cx, by - sh - 4), 10, fruit)
	draw_circle(Vector2(cx - 3, by - sh - 7), 3, Color(1, 1, 1, 0.35))

func _draw_plant_growing(cx: float, cy: float, leaf: Color, prog: float):
	var by: float = cy
	var sh: float = 10.0 + prog * 18.0
	draw_line(Vector2(cx, by), Vector2(cx, by - sh), Color(0.2, 0.5, 0.12), 2.5)
	draw_circle(Vector2(cx, by - sh), 5 + int(prog * 4), leaf)
	draw_circle(Vector2(cx - 6, by - sh * 0.5), 4, leaf)
	draw_circle(Vector2(cx + 6, by - sh * 0.45), 3.5, leaf)

func _draw_plant_seed(cx: float, cy: float, prog: float):
	var by: float = cy
	# Small sprout
	var h: float = 3.0 + prog * 10.0
	draw_line(Vector2(cx, by), Vector2(cx, by - h), Color(0.25, 0.6, 0.15), 2.0)
	if prog > 0.05:
		draw_circle(Vector2(cx - 2, by - h), 3, Color(0.3, 0.75, 0.2))
		draw_circle(Vector2(cx + 2, by - h + 1), 2.5, Color(0.3, 0.75, 0.2))
	# Seed
	draw_circle(Vector2(cx, by + 2), 3.5, Color(0.6, 0.45, 0.25))

func _draw_land_tile(corners: PackedVector2Array, cell: Dictionary):
	var texture := _get_land_texture(cell)
	if texture == null:
		draw_colored_polygon(corners, Color(0.52, 0.36, 0.20))
		return
	var size := texture.get_size()
	if size.x <= 0.0 or size.y <= 0.0:
		return
	var center := Vector2.ZERO
	for point in corners:
		center += point
	center /= maxf(float(corners.size()), 1.0)
	var source := _get_land_texture_source_rect(cell, size)
	var target_w := TW
	var target_h := TH
	if not _is_cell_unlocked(cell):
		target_w = TW
		target_h = TH
	var dest := Rect2(center.x - target_w * 0.5, center.y - target_h * 0.5, target_w, target_h)
	draw_texture_rect_region(texture, dest, source)

func _draw_land_tooltip(col: int, row: int):
	var cell: Dictionary = farm[row][col]
	var sp := _get_plot_position(col, row)
	var tw := 176.0
	var th := 78.0
	var tx := clampf(sp.x - tw * 0.5, 5.0, VIEW_W - tw - 5.0)
	var ty := clampf(sp.y - TH * 0.5 - th - 18.0, 5.0, VIEW_H - th - 5.0)
	var land_level := int(cell.get("land_level", 1))
	var work := int(cell.get("land_work", 0))
	draw_rect(Rect2(tx, ty, tw, th), Color(0.08, 0.05, 0.02, 0.92))
	draw_rect(Rect2(tx, ty, tw, th), Color(0.55, 0.42, 0.2), false, 2)
	draw_rect(Rect2(tx, ty, tw, 22), Color(0.38, 0.28, 0.12))
	_draw_text(tx + 8, ty + 3, _get_land_level_name(land_level), 13, Color(1.0, 0.92, 0.72))
	if land_level >= LAND_LEVEL_MAX:
		_draw_text(tx + 10, ty + 32, "已是最高等级", 12, Color(0.95, 0.82, 0.42))
	else:
		_draw_text(tx + 10, ty + 32, "土地经验: " + str(work) + "/" + str(LAND_UPGRADE_WORK_REQUIRED), 12, Color(0.95, 0.82, 0.42))
		_draw_text(tx + 10, ty + 50, "收获后增加 1 点", 11, Color(0.7, 0.85, 1.0))
	var arrow_x: float = clampf(sp.x, tx + 10.0, tx + tw - 10.0)
	var arrow_pts: PackedVector2Array = PackedVector2Array([
		Vector2(arrow_x - 6.0, ty + th),
		Vector2(arrow_x, ty + th + 8.0),
		Vector2(arrow_x + 6.0, ty + th),
	])
	draw_colored_polygon(arrow_pts, Color(0.08, 0.05, 0.02, 0.92))

func _get_land_texture_source_rect(cell: Dictionary, size: Vector2) -> Rect2:
	var key := _get_land_texture_key(cell)
	if land_texture_source_rects.has(key):
		return land_texture_source_rects[key]
	return Rect2(Vector2.ZERO, size)

func _get_texture_alpha_bounds(texture: Texture2D) -> Rect2:
	var image := texture.get_image()
	if image == null or image.is_empty():
		return Rect2(Vector2.ZERO, texture.get_size())
	var min_x := image.get_width()
	var min_y := image.get_height()
	var max_x := -1
	var max_y := -1
	for y in range(image.get_height()):
		for x in range(image.get_width()):
			if image.get_pixel(x, y).a > 0.01:
				min_x = mini(min_x, x)
				min_y = mini(min_y, y)
				max_x = maxi(max_x, x)
				max_y = maxi(max_y, y)
	if max_x < min_x or max_y < min_y:
		return Rect2(Vector2.ZERO, image.get_size())
	return Rect2(min_x, min_y, max_x - min_x + 1, max_y - min_y + 1)

func _inset_polygon(points: PackedVector2Array, amount: float) -> PackedVector2Array:
	var center := Vector2.ZERO
	for point in points:
		center += point
	center /= maxf(float(points.size()), 1.0)
	var result := PackedVector2Array()
	for point in points:
		result.append(center + (point - center) * (1.0 - amount))
	return result

func _get_land_texture(cell: Dictionary) -> Texture2D:
	return land_textures.get(_get_land_texture_key(cell), null) as Texture2D

func _get_land_texture_key(cell: Dictionary) -> String:
	if not _is_cell_unlocked(cell):
		return "locked"
	var level := _get_land_level_key(int(cell.get("land_level", 1)))
	var suffix := "wet" if float(cell.get("wet_timer", 0.0)) > 0.0 else "dry"
	return level + "_" + suffix

func _get_land_level_key(land_level: int) -> String:
	if land_level <= 2:
		return "yellow"
	if land_level == 3:
		return "red"
	return "black"

func _get_crop_stage_texture(cid: int, prog: float) -> Texture2D:
	var stage := _get_growth_stage(prog)
	if stage < 0:
		return null
	return CropAtlas.get_stage_texture(str(CROPS[cid][4]), stage + 1)

func _get_crop_seed_texture(cid: int) -> Texture2D:
	return CropAtlas.get_stage_texture(str(CROPS[cid][4]), 0)

func _get_crop_mature_texture(cid: int) -> Texture2D:
	return CropAtlas.get_stage_texture(str(CROPS[cid][4]), 3)

func _get_growth_stage(prog: float) -> int:
	if prog < 0.18:
		return -1
	if prog < 0.45:
		return 0
	if prog < 0.72:
		return 1
	if prog < 0.9:
		return 2
	return 2

func _draw_crop_atlas_texture(cx: float, tile_center_y: float, texture: Texture2D, prog: float, cid: int, stage: int):
	var size := texture.get_size()
	if size.x <= 0.0 or size.y <= 0.0:
		return
	var scale_factor := _get_crop_scale(cid, size, stage)
	var draw_size := size * scale_factor
	var draw_pos := Vector2(cx - draw_size.x * 0.5, tile_center_y - draw_size.y)
	draw_texture_rect(texture, Rect2(draw_pos, draw_size), false)

func _get_crop_scale(cid: int, size: Vector2, stage: int) -> float:
	var target_size := 110.0
	var crop_key := str(CROPS[cid][4])
	match crop_key:
		"lettuce":
			target_size = 92.0
		"pepper":
			target_size = 100.0
		"eggplant":
			target_size = 102.0
		"tomato":
			target_size = 104.0
		"strawberry":
			target_size = 96.0
		"corn":
			target_size = 118.0
		"sunflower":
			target_size = 126.0
		"pumpkin":
			target_size = 112.0
		"watermelon":
			target_size = 114.0
	var stage_multiplier := 1.0
	match stage:
		0:
			stage_multiplier = 0.62
		1:
			stage_multiplier = 0.82
		2:
			stage_multiplier = 1.0
	var base_scale := minf(1.0, target_size / maxf(size.x, size.y)) * stage_multiplier
	var width_limit_scale := (TW * 0.8) / maxf(size.x, 1.0)
	return minf(base_scale, width_limit_scale)

func _get_crop_ground_offset(cid: int) -> float:
	var crop_key := str(CROPS[cid][4])
	match crop_key:
		"lettuce":
			return -8.0
		"pepper":
			return -10.0
		"eggplant":
			return -9.0
		"tomato":
			return -10.0
		"strawberry":
			return -8.0
		"corn":
			return -12.0
		"sunflower":
			return -14.0
		"pumpkin":
			return -6.0
		"watermelon":
			return -6.0
		_:
			return -10.0

func _draw_crop_preview(x: float, y: float, texture: Texture2D, label: String):
	var bg_rect := Rect2(x, y, 120, 92)
	draw_rect(bg_rect, Color(0.14, 0.10, 0.06, 0.75))
	draw_rect(bg_rect, Color(0.40, 0.30, 0.16), false, 2)
	if texture != null:
		var size := texture.get_size()
		if size.x > 0.0 and size.y > 0.0:
			var scale_factor := minf(0.42, 62.0 / maxf(size.x, size.y))
			var draw_size := size * scale_factor
			var draw_pos := Vector2(x + 60 - draw_size.x * 0.5, y + 60 - draw_size.y)
			draw_texture_rect(texture, Rect2(draw_pos, draw_size), false)
	else:
		draw_circle(Vector2(x + 60, y + 42), 16, Color(0.35, 0.42, 0.26))
	_draw_text(x + 34, y + 68, label, 12, Color(1, 0.95, 0.85))

func _draw_ui_seed_thumbnail(rect: Rect2, texture: Texture2D):
	draw_rect(rect, Color(1, 1, 1, 0.22))
	draw_rect(rect, Color(0.42, 0.30, 0.16, 0.55), false, 1.0)
	if texture == null:
		return
	var size := texture.get_size()
	if size.x <= 0.0 or size.y <= 0.0:
		return
	var scale_factor := minf(rect.size.x / size.x, rect.size.y / size.y)
	var draw_size := size * scale_factor
	var draw_pos := rect.position + (rect.size - draw_size) * 0.5
	draw_texture_rect(texture, Rect2(draw_pos, draw_size), false)

func _draw_seed_preview_texture(cx: float, cy: float, texture: Texture2D):
	var size := texture.get_size()
	if size.x <= 0.0 or size.y <= 0.0:
		return
	var target := 26.0
	var scale_factor := minf(target / size.x, target / size.y)
	var draw_size := size * scale_factor
	var draw_pos := Vector2(cx - draw_size.x * 0.5, cy - draw_size.y * 0.5)
	draw_texture_rect(texture, Rect2(draw_pos, draw_size), false)

# ---- Text helper ----
func _draw_text(x: float, y: float, text: String, size: int, color: Color):
	var font: Font = ThemeDB.fallback_font
	draw_string(font, Vector2(x, y + float(size) * 0.8), text, HORIZONTAL_ALIGNMENT_LEFT, -1, size, color)
