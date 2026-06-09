@tool
extends Node2D

const HALF_W := 84.0
const HALF_H := 42.0
const LAND_TEXTURE_PATH := "res://assets/land/land_yellow_dry.png"

@export var label_text := ""

var _land_texture: Texture2D = null

func _ready() -> void:
	_sync_visual()
	queue_redraw()

func _process(_delta: float) -> void:
	_sync_visual()
	if Engine.is_editor_hint():
		queue_redraw()

func _draw() -> void:
	if not Engine.is_editor_hint():
		return
	var corners := PackedVector2Array([
		Vector2(0.0, -HALF_H),
		Vector2(HALF_W, 0.0),
		Vector2(0.0, HALF_H),
		Vector2(-HALF_W, 0.0),
	])
	draw_polyline(
		PackedVector2Array([corners[0], corners[1], corners[2], corners[3], corners[0]]),
		Color(0.16, 0.95, 1.0, 0.95),
		2.0
	)
	draw_circle(Vector2.ZERO, 4.0, Color(1.0, 0.45, 0.08, 0.95))
	if label_text != "":
		var font := ThemeDB.fallback_font
		var label_size := font.get_string_size(label_text, HORIZONTAL_ALIGNMENT_LEFT, -1, 14)
		draw_rect(
			Rect2(-label_size.x * 0.5 - 4.0, -10.0, label_size.x + 8.0, 18.0),
			Color(0.05, 0.08, 0.10, 0.72)
		)
		draw_string(
			font,
			Vector2(-label_size.x * 0.5, 4.0),
			label_text,
			HORIZONTAL_ALIGNMENT_LEFT,
			-1,
			14,
			Color(1.0, 1.0, 1.0, 0.95)
		)

func _sync_visual() -> void:
	if _land_texture == null:
		_land_texture = load(LAND_TEXTURE_PATH) as Texture2D
	var visual := get_node_or_null("Visual") as Polygon2D
	if visual == null:
		visual = Polygon2D.new()
		visual.name = "Visual"
		add_child(visual)
		if owner != null:
			visual.owner = owner
	visual.show_behind_parent = false
	visual.visible = Engine.is_editor_hint()
	visual.z_index = -5
	visual.color = Color(1.0, 1.0, 1.0, 0.92)
	visual.texture = _land_texture
	visual.polygon = PackedVector2Array([
		Vector2(0.0, -HALF_H),
		Vector2(HALF_W, 0.0),
		Vector2(0.0, HALF_H),
		Vector2(-HALF_W, 0.0),
	])
	if _land_texture != null:
		var size := _land_texture.get_size()
		visual.uv = PackedVector2Array([
			Vector2(size.x * 0.5, 0.0),
			Vector2(size.x, size.y * 0.5),
			Vector2(size.x * 0.5, size.y),
			Vector2(0.0, size.y * 0.5),
		])
	var tint := get_node_or_null("Tint") as Polygon2D
	if tint == null:
		tint = Polygon2D.new()
		tint.name = "Tint"
		add_child(tint)
		if owner != null:
			tint.owner = owner
	tint.show_behind_parent = false
	tint.visible = Engine.is_editor_hint()
	tint.z_index = -4
	tint.texture = null
	tint.color = Color(0.58, 0.36, 0.16, 0.28)
	tint.polygon = visual.polygon
	var outline := get_node_or_null("Outline") as Line2D
	if outline == null:
		outline = Line2D.new()
		outline.name = "Outline"
		add_child(outline)
		if owner != null:
			outline.owner = owner
	outline.show_behind_parent = false
	outline.visible = Engine.is_editor_hint()
	outline.z_index = -3
	outline.width = 2.0
	outline.default_color = Color(0.38, 0.22, 0.10, 0.95)
	outline.closed = true
	outline.points = visual.polygon
