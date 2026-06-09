extends Node2D

const LEAF_TEXTURE: Texture2D = preload("res://assets/plants/source/opengameart/farm_sprites/Seeds2_0.png")
const REGION_LOWER_LEFT: Rect2 = Rect2(50, 46, 11, 27)
const REGION_LOWER_RIGHT: Rect2 = Rect2(82, 70, 13, 23)
const REGION_UPPER_LEFT: Rect2 = Rect2(128, 49, 11, 23)
const REGION_UPPER_RIGHT: Rect2 = Rect2(183, 70, 11, 20)
const BASE_SEGMENT: float = 30.0
const MID_SEGMENT: float = 26.0
const TOP_SEGMENT: float = 22.0

@export_range(0.25, 1.0, 0.01) var growth: float = 1.0
@export var wind_speed: float = 1.4
@export var wind_amount_deg: float = 10.0
@export var stem_color: Color = Color(0.20, 0.56, 0.18)
@export var fruit_color: Color = Color(0.86, 0.23, 0.18)
@export var fruit_shadow: Color = Color(0.57, 0.10, 0.09)

@onready var stem_base: Node2D = $StemBase
@onready var stem_mid: Node2D = $StemBase/StemMid
@onready var stem_top: Node2D = $StemBase/StemMid/StemTop
@onready var leaf_left_lower: Sprite2D = $StemBase/StemMid/LeafLeftLower
@onready var leaf_right_lower: Sprite2D = $StemBase/StemMid/LeafRightLower
@onready var leaf_left_upper: Sprite2D = $StemBase/StemMid/StemTop/LeafLeftUpper
@onready var leaf_right_upper: Sprite2D = $StemBase/StemMid/StemTop/LeafRightUpper
@onready var fruit_left: Node2D = $StemBase/StemMid/StemTop/FruitLeft
@onready var fruit_right: Node2D = $StemBase/StemMid/StemTop/FruitRight

var sway_time: float = 0.0

func _ready() -> void:
	_setup_leaf(leaf_left_lower, REGION_LOWER_LEFT, Vector2(-17.0, -5.0), -122.0, Vector2(-3.0, 3.0))
	_setup_leaf(leaf_right_lower, REGION_LOWER_RIGHT, Vector2(17.0, -3.0), 110.0, Vector2(3.0, 3.0))
	_setup_leaf(leaf_left_upper, REGION_UPPER_LEFT, Vector2(-11.0, -4.0), -136.0, Vector2(-2.8, 2.8))
	_setup_leaf(leaf_right_upper, REGION_UPPER_RIGHT, Vector2(11.0, -1.0), 118.0, Vector2(2.8, 2.8))
	fruit_left.position = Vector2(-8.0, 10.0)
	fruit_right.position = Vector2(10.0, 7.0)
	queue_redraw()

func _process(delta: float) -> void:
	sway_time += delta
	var primary_sway: float = sin(sway_time * wind_speed)
	var secondary_sway: float = sin(sway_time * wind_speed * 0.63 + 1.4)
	var growth_blend: float = growth
	stem_base.rotation = deg_to_rad(primary_sway * wind_amount_deg * 0.22 * growth_blend)
	stem_mid.rotation = deg_to_rad(primary_sway * wind_amount_deg * 0.36 * growth_blend + secondary_sway * 1.4)
	stem_top.rotation = deg_to_rad(primary_sway * wind_amount_deg * 0.58 * growth_blend + secondary_sway * 2.1)
	leaf_left_lower.rotation = deg_to_rad(-122.0 + primary_sway * 6.0)
	leaf_right_lower.rotation = deg_to_rad(110.0 + primary_sway * 5.0)
	leaf_left_upper.rotation = deg_to_rad(-136.0 + primary_sway * 8.0 + secondary_sway * 3.0)
	leaf_right_upper.rotation = deg_to_rad(118.0 + primary_sway * 7.0 + secondary_sway * 2.0)
	var fruit_growth: float = clampf((growth - 0.58) / 0.42, 0.0, 1.0)
	var fruit_bob: float = sin(sway_time * wind_speed * 1.8) * fruit_growth
	fruit_left.position = Vector2(-8.0, 10.0 + fruit_bob * 2.0)
	fruit_right.position = Vector2(10.0, 7.0 - fruit_bob * 1.6)
	_update_leaf_visibility(growth_blend)
	queue_redraw()

func _draw() -> void:
	var base_start: Vector2 = Vector2.ZERO
	var mid_start: Vector2 = to_local(stem_mid.global_position)
	var top_start: Vector2 = to_local(stem_top.global_position)
	var base_end: Vector2 = mid_start
	var mid_end: Vector2 = top_start
	var top_end: Vector2 = top_start + Vector2(0.0, -TOP_SEGMENT * growth).rotated(stem_top.global_rotation)
	_draw_segment(base_start, base_end, 10.0 * growth, stem_color)
	_draw_segment(mid_start, mid_end, 7.0 * growth, stem_color.darkened(0.06))
	_draw_segment(mid_end, top_end, 5.0 * growth, stem_color.darkened(0.12))
	_draw_fruit_cluster()

func _setup_leaf(sprite: Sprite2D, region_rect: Rect2, local_pos: Vector2, rotation_deg: float, scale_value: Vector2) -> void:
	sprite.texture = LEAF_TEXTURE
	sprite.region_enabled = true
	sprite.region_rect = region_rect
	sprite.centered = true
	sprite.position = local_pos
	sprite.rotation_degrees = rotation_deg
	sprite.scale = scale_value
	sprite.offset = Vector2(0.0, region_rect.size.y * 0.18)
	if sprite.scale.x < 0.0:
		sprite.flip_h = true
	else:
		sprite.flip_h = false

func _update_leaf_visibility(growth_blend: float) -> void:
	var lower_scale: float = lerpf(1.4, 3.0, growth_blend)
	var upper_scale: float = lerpf(1.2, 2.8, growth_blend)
	leaf_left_lower.scale = Vector2(-lower_scale, lower_scale)
	leaf_right_lower.scale = Vector2(lower_scale, lower_scale)
	leaf_left_upper.scale = Vector2(-upper_scale, upper_scale)
	leaf_right_upper.scale = Vector2(upper_scale, upper_scale)
	var alpha_value: float = clampf(0.2 + growth_blend * 0.8, 0.0, 1.0)
	leaf_left_lower.modulate.a = alpha_value
	leaf_right_lower.modulate.a = alpha_value
	leaf_left_upper.modulate.a = alpha_value
	leaf_right_upper.modulate.a = alpha_value

func _draw_segment(from_point: Vector2, to_point: Vector2, width: float, color_value: Color) -> void:
	draw_line(from_point, to_point, color_value, width, true)
	draw_line(from_point + Vector2(1.0, 0.0), to_point + Vector2(1.0, 0.0), color_value.lightened(0.14), maxf(1.0, width - 3.0), true)

func _draw_fruit_cluster() -> void:
	var fruit_growth: float = clampf((growth - 0.58) / 0.42, 0.0, 1.0)
	if fruit_growth <= 0.0:
		return
	var left_pos: Vector2 = to_local(fruit_left.global_position)
	var right_pos: Vector2 = to_local(fruit_right.global_position)
	var radius_big: float = lerpf(1.0, 9.0, fruit_growth)
	var radius_small: float = lerpf(1.0, 7.0, fruit_growth)
	draw_circle(left_pos + Vector2(1.0, 2.0), radius_big, fruit_shadow)
	draw_circle(right_pos + Vector2(1.0, 2.0), radius_small, fruit_shadow)
	draw_circle(left_pos, radius_big, fruit_color)
	draw_circle(right_pos, radius_small, fruit_color)
	draw_circle(left_pos + Vector2(-2.0, -3.0), radius_big * 0.23, Color(1.0, 1.0, 1.0, 0.4))
	draw_circle(right_pos + Vector2(-1.0, -2.0), radius_small * 0.25, Color(1.0, 1.0, 1.0, 0.35))
