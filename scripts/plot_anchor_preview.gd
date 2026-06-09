@tool
extends Node2D

const COLS := 6
const ROWS := 5
const HALF_W := 84.0
const HALF_H := 42.0
const ANCHORS_PATH := "../PlotAnchors"

func _process(_delta: float) -> void:
	if Engine.is_editor_hint():
		queue_redraw()

func _draw() -> void:
	if not Engine.is_editor_hint():
		return
	var anchors := get_node_or_null(ANCHORS_PATH)
	if anchors == null:
		return
	var font := ThemeDB.fallback_font
	for r in range(ROWS):
		for c in range(COLS):
			var node := anchors.get_node_or_null("Plot_%d_%d" % [r, c])
			if not (node is Node2D):
				continue
			var center := (node as Node2D).position
			var corners := PackedVector2Array([
				center + Vector2(0.0, -HALF_H),
				center + Vector2(HALF_W, 0.0),
				center + Vector2(0.0, HALF_H),
				center + Vector2(-HALF_W, 0.0),
			])
			draw_colored_polygon(corners, Color(0.1, 0.9, 1.0, 0.08))
			for i in range(4):
				draw_line(corners[i], corners[(i + 1) % 4], Color(0.1, 0.9, 1.0, 0.8), 2.0)
			draw_circle(center, 4.0, Color(1.0, 0.4, 0.1, 0.95))
			var label := "%d,%d" % [r, c]
			var label_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, 14)
			draw_rect(
				Rect2(center.x - label_size.x * 0.5 - 4.0, center.y - 12.0, label_size.x + 8.0, 18.0),
				Color(0.04, 0.08, 0.12, 0.72)
			)
			draw_string(
				font,
				Vector2(center.x - label_size.x * 0.5, center.y + 2.0),
				label,
				HORIZONTAL_ALIGNMENT_LEFT,
				-1,
				14,
				Color(1.0, 1.0, 1.0, 0.95)
			)
