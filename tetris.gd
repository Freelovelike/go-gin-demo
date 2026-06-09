extends Node2D

const COLS := 10
const ROWS := 20
const CELL := 30
const BOARD_X := 50
const BOARD_Y := 50

var SHAPES := {}
var COLORS := {}

var board: Array = []
var current_piece := ""
var current_rot := 0
var current_pos_x := 0
var current_pos_y := 0
var next_piece := ""
var score := 0
var level := 1
var lines_cleared := 0
var game_over := false
var drop_timer := 0.0
var drop_interval := 0.5
var paused := false
var bag: Array = []

func _ready():
	_init_shapes()
	_init_game()

func _init_shapes():
	SHAPES = {
		"I": [[[0,0],[1,0],[2,0],[3,0]],[[0,0],[0,1],[0,2],[0,3]],[[0,0],[1,0],[2,0],[3,0]],[[0,0],[0,1],[0,2],[0,3]]],
		"O": [[[0,0],[1,0],[0,1],[1,1]],[[0,0],[1,0],[0,1],[1,1]],[[0,0],[1,0],[0,1],[1,1]],[[0,0],[1,0],[0,1],[1,1]]],
		"T": [[[0,0],[1,0],[2,0],[1,1]],[[0,0],[0,1],[0,2],[1,1]],[[0,1],[1,1],[2,1],[1,0]],[[1,0],[1,1],[1,2],[0,1]]],
		"S": [[[1,0],[2,0],[0,1],[1,1]],[[0,0],[0,1],[1,1],[1,2]],[[1,0],[2,0],[0,1],[1,1]],[[0,0],[0,1],[1,1],[1,2]]],
		"Z": [[[0,0],[1,0],[1,1],[2,1]],[[1,0],[0,1],[1,1],[0,2]],[[0,0],[1,0],[1,1],[2,1]],[[1,0],[0,1],[1,1],[0,2]]],
		"J": [[[0,0],[0,1],[1,1],[2,1]],[[0,0],[1,0],[0,1],[0,2]],[[0,0],[1,0],[2,0],[2,1]],[[1,0],[1,1],[0,2],[1,2]]],
		"L": [[[2,0],[0,1],[1,1],[2,1]],[[0,0],[0,1],[0,2],[1,2]],[[0,0],[1,0],[2,0],[0,1]],[[0,0],[1,0],[1,1],[1,2]]],
	}
	COLORS = {
		"I": Color(0.0, 1.0, 1.0),
		"O": Color(1.0, 1.0, 0.0),
		"T": Color(0.6, 0.0, 1.0),
		"S": Color(0.0, 1.0, 0.0),
		"Z": Color(1.0, 0.0, 0.0),
		"J": Color(0.0, 0.0, 1.0),
		"L": Color(1.0, 0.5, 0.0),
	}

func _init_game():
	board = []
	for _i in range(ROWS):
		var row: Array = []
		for _j in range(COLS):
			row.append("")
		board.append(row)
	score = 0
	level = 1
	lines_cleared = 0
	game_over = false
	paused = false
	drop_interval = 0.5
	bag = []
	current_piece = ""
	next_piece = ""
	_spawn_piece()
	queue_redraw()

# ---- 随机袋 ----
func _fill_bag():
	bag = ["I","O","T","S","Z","J","L"]
	bag.shuffle()

func _next_from_bag() -> String:
	if bag.is_empty():
		_fill_bag()
	return str(bag.pop_back())

func _spawn_piece():
	current_piece = next_piece if next_piece != "" else _next_from_bag()
	next_piece = _next_from_bag()
	current_rot = 0
	var shape: Array = SHAPES[current_piece][current_rot]
	var min_x := 99
	var max_x := -1
	for cell in shape:
		var cx: int = int(cell[0])
		min_x = mini(min_x, cx)
		max_x = maxi(max_x, cx)
	var piece_w := max_x - min_x + 1
	current_pos_x = (COLS - piece_w) / 2 - min_x
	current_pos_y = 0
	if _collides(current_rot, current_pos_x, current_pos_y):
		game_over = true
		queue_redraw()

# ---- 碰撞 ----
func _collides(rot: int, px: int, py: int) -> bool:
	var shape: Array = SHAPES[current_piece][rot % 4]
	for cell in shape:
		var c: int = int(cell[0]) + px
		var r: int = int(cell[1]) + py
		if c < 0 or c >= COLS or r >= ROWS:
			return true
		if r >= 0 and board[r][c] != "":
			return true
	return false

func _lock_piece():
	var shape: Array = SHAPES[current_piece][current_rot]
	for cell in shape:
		var c: int = int(cell[0]) + current_pos_x
		var r: int = int(cell[1]) + current_pos_y
		if r >= 0 and r < ROWS and c >= 0 and c < COLS:
			board[r][c] = current_piece
	_clear_lines()
	_spawn_piece()

func _clear_lines():
	var cleared := 0
	var r := ROWS - 1
	while r >= 0:
		var full := true
		for c in range(COLS):
			if board[r][c] == "":
				full = false
				break
		if full:
			board.remove_at(r)
			var new_row: Array = []
			for _j in range(COLS):
				new_row.append("")
			board.insert(0, new_row)
			cleared += 1
		else:
			r -= 1
	if cleared > 0:
		var points: Array = [0, 100, 300, 500, 800]
		score += int(points[mini(cleared, 4)]) * level
		lines_cleared += cleared
		level = lines_cleared / 10 + 1
		drop_interval = maxf(0.05, 0.5 - (level - 1) * 0.05)
		queue_redraw()

# ---- 移动 ----
func _try_move(dx: int, dy: int) -> bool:
	if not _collides(current_rot, current_pos_x + dx, current_pos_y + dy):
		current_pos_x += dx
		current_pos_y += dy
		queue_redraw()
		return true
	return false

func _try_rotate() -> bool:
	var new_rot := (current_rot + 1) % 4
	var kicks: Array = [[0,0],[-1,0],[1,0],[0,-1],[-1,-1],[1,-1]]
	for kick in kicks:
		var kx: int = int(kick[0])
		var ky: int = int(kick[1])
		if not _collides(new_rot, current_pos_x + kx, current_pos_y + ky):
			current_rot = new_rot
			current_pos_x += kx
			current_pos_y += ky
			queue_redraw()
			return true
	return false

func _hard_drop():
	while _try_move(0, 1):
		pass
	_lock_piece()

func _ghost_distance() -> int:
	var dist := 0
	while not _collides(current_rot, current_pos_x, current_pos_y + dist + 1):
		dist += 1
	return dist

# ---- 输入 ----
func _unhandled_input(event: InputEvent):
	if event is InputEventKey and event.pressed:
		if event.keycode == KEY_R:
			_init_game()
			return
		if game_over:
			return
		if event.keycode == KEY_P:
			paused = !paused
			queue_redraw()
			return
		if paused:
			return
		match event.keycode:
			KEY_LEFT:
				_try_move(-1, 0)
			KEY_RIGHT:
				_try_move(1, 0)
			KEY_DOWN:
				if _try_move(0, 1):
					score += 1
					drop_timer = 0.0
			KEY_UP:
				_try_rotate()
			KEY_SPACE:
				_hard_drop()

# ---- 游戏循环 ----
func _process(delta: float):
	if game_over or paused:
		return
	drop_timer += delta
	if drop_timer >= drop_interval:
		drop_timer = 0.0
		if not _try_move(0, 1):
			_lock_piece()
		queue_redraw()

# ---- 绘制 ----
func _draw():
	_draw_background()
	_draw_board()
	_draw_ghost()
	_draw_current_piece()
	_draw_side_panel()

func _draw_background():
	draw_rect(Rect2(0, 0, 1280, 720), Color(0.08, 0.08, 0.12))

func _draw_board():
	draw_rect(Rect2(BOARD_X - 3, BOARD_Y - 3, COLS * CELL + 6, ROWS * CELL + 6), Color(0.4, 0.4, 0.5), false, 2.0)
	draw_rect(Rect2(BOARD_X, BOARD_Y, COLS * CELL, ROWS * CELL), Color(0.05, 0.05, 0.08))
	for r in range(ROWS + 1):
		draw_line(Vector2(BOARD_X, BOARD_Y + r * CELL), Vector2(BOARD_X + COLS * CELL, BOARD_Y + r * CELL), Color(0.12, 0.12, 0.18), 1.0)
	for c in range(COLS + 1):
		draw_line(Vector2(BOARD_X + c * CELL, BOARD_Y), Vector2(BOARD_X + c * CELL, BOARD_Y + ROWS * CELL), Color(0.12, 0.12, 0.18), 1.0)
	for r in range(ROWS):
		for c in range(COLS):
			if board[r][c] != "":
				_draw_cell(c, r, COLORS[str(board[r][c])])

func _draw_current_piece():
	if current_piece == "":
		return
	var shape: Array = SHAPES[current_piece][current_rot]
	for cell in shape:
		var c: int = int(cell[0]) + current_pos_x
		var r: int = int(cell[1]) + current_pos_y
		if r >= 0:
			_draw_cell(c, r, COLORS[current_piece])

func _draw_ghost():
	if current_piece == "":
		return
	var dist := _ghost_distance()
	if dist == 0:
		return
	var col: Color = COLORS[current_piece]
	var ghost_col := Color(col.r, col.g, col.b, 0.2)
	var shape: Array = SHAPES[current_piece][current_rot]
	for cell in shape:
		var c: int = int(cell[0]) + current_pos_x
		var r: int = int(cell[1]) + current_pos_y + dist
		if r >= 0 and r < ROWS:
			draw_rect(Rect2(BOARD_X + c * CELL + 1, BOARD_Y + r * CELL + 1, CELL - 2, CELL - 2), ghost_col)

func _draw_cell(col: int, row: int, color: Color):
	var x := BOARD_X + col * CELL
	var y := BOARD_Y + row * CELL
	draw_rect(Rect2(x + 1, y + 1, CELL - 2, CELL - 2), color)
	draw_line(Vector2(x + 1, y + 1), Vector2(x + CELL - 2, y + 1), Color(1, 1, 1, 0.3), 2.0)
	draw_line(Vector2(x + 1, y + 1), Vector2(x + 1, y + CELL - 2), Color(1, 1, 1, 0.3), 2.0)
	draw_line(Vector2(x + CELL - 2, y + 1), Vector2(x + CELL - 2, y + CELL - 2), Color(0, 0, 0, 0.3), 2.0)
	draw_line(Vector2(x + 1, y + CELL - 2), Vector2(x + CELL - 2, y + CELL - 2), Color(0, 0, 0, 0.3), 2.0)

func _draw_side_panel():
	var px := BOARD_X + COLS * CELL + 40
	var py := BOARD_Y
	draw_rect(Rect2(px - 10, py - 10, 250, 650), Color(0.1, 0.1, 0.15))
	draw_rect(Rect2(px - 12, py - 12, 254, 654), Color(0.3, 0.3, 0.4), false, 2.0)
	_draw_label(px, py, "SCORE")
	_draw_number(px, py + 30, score)
	_draw_label(px, py + 80, "LEVEL")
	_draw_number(px, py + 110, level)
	_draw_label(px, py + 160, "LINES")
	_draw_number(px, py + 190, lines_cleared)
	_draw_label(px, py + 250, "NEXT")
	if next_piece != "":
		var shape: Array = SHAPES[next_piece][0]
		for cell in shape:
			var cx: int = int(cell[0]) * (CELL - 4) + px
			var cy: int = int(cell[1]) * (CELL - 4) + py + 290
			draw_rect(Rect2(cx, cy, CELL - 6, CELL - 6), COLORS[next_piece])
	if game_over:
		_draw_label(px, py + 400, "GAME OVER!")
		_draw_label(px, py + 420, "PRESS R")
	elif paused:
		_draw_label(px, py + 400, "PAUSED")
	_draw_label(px, py + 480, "CONTROLS")
	_draw_label(px, py + 510, "<- -> MOVE")
	_draw_label(px, py + 530, "UP   ROTATE")
	_draw_label(px, py + 550, "DOWN SOFT DROP")
	_draw_label(px, py + 570, "SPACE HARD DROP")
	_draw_label(px, py + 590, "P  PAUSE")
	_draw_label(px, py + 610, "R  RESTART")

func _draw_label(x: int, y: int, text: String):
	draw_rect(Rect2(x - 2, y - 2, text.length() * 12 + 4, 18), Color(0.08, 0.08, 0.12))

func _draw_number(x: int, y: int, num: int):
	var str_num := str(num)
	for i in range(str_num.length()):
		var digit: int = int(str_num[i])
		_draw_digit(x + i * 16, y, digit)

func _draw_digit(x: int, y: int, digit: int):
	var patterns: Array = [
		[1,1,1,1,0,1,1,0,1,1,0,1,1,1,1],
		[0,1,0,1,1,0,0,1,0,0,1,0,1,1,1],
		[1,1,1,0,0,1,1,1,1,1,0,0,1,1,1],
		[1,1,1,0,0,1,1,1,1,0,0,1,1,1,1],
		[1,0,1,1,0,1,1,1,1,0,0,1,0,0,1],
		[1,1,1,1,0,0,1,1,1,0,0,1,1,1,1],
		[1,1,1,1,0,0,1,1,1,1,0,1,1,1,1],
		[1,1,1,0,0,1,0,1,0,0,1,0,0,1,0],
		[1,1,1,1,0,1,1,1,1,1,0,1,1,1,1],
		[1,1,1,1,0,1,1,1,1,0,0,1,1,1,1],
	]
	if digit < 0 or digit > 9:
		return
	var p: Array = patterns[digit]
	for row in range(5):
		for col in range(3):
			if int(p[row * 3 + col]) == 1:
				draw_rect(Rect2(x + col * 5, y + row * 5, 4, 4), Color(0.9, 0.9, 0.9))
