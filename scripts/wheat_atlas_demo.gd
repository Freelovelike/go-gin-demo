extends Node2D

const CropAtlas = preload("res://scripts/crop_atlas.gd")

var stage_sprites: Array[Sprite2D] = []

func _ready() -> void:
	for i in range(4):
		var sprite := Sprite2D.new()
		sprite.texture = CropAtlas.get_wheat_stage_texture(i)
		sprite.centered = true
		sprite.position = Vector2(180 + i * 250, 240)
		add_child(sprite)
		stage_sprites.append(sprite)

	queue_redraw()

func _draw() -> void:
	draw_rect(Rect2(0, 0, 1280, 720), Color(0.95, 0.95, 0.92))
	draw_string(ThemeDB.fallback_font, Vector2(40, 56), "Wheat Atlas Demo", HORIZONTAL_ALIGNMENT_LEFT, -1, 28, Color(0.2, 0.2, 0.16))
	for i in range(4):
		draw_string(
			ThemeDB.fallback_font,
			Vector2(120 + i * 250, 430),
			"Stage " + str(i + 1),
			HORIZONTAL_ALIGNMENT_LEFT,
			-1,
			20,
			Color(0.32, 0.26, 0.18)
		)
