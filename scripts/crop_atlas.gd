extends RefCounted
class_name CropAtlas

const IMPORTED_CROP_STAGE_PATHS := {
	"lettuce": [
		"res://assets/plants/imported/lettuce/stage_1.png",
		"res://assets/plants/imported/lettuce/stage_2.png",
		"res://assets/plants/imported/lettuce/stage_3.png",
		"res://assets/plants/imported/lettuce/stage_4.png",
	],
	"pepper": [
		"res://assets/plants/imported/pepper/stage_1.png",
		"res://assets/plants/imported/pepper/stage_2.png",
		"res://assets/plants/imported/pepper/stage_3.png",
		"res://assets/plants/imported/pepper/stage_4.png",
	],
	"eggplant": [
		"res://assets/plants/imported/eggplant/stage_1.png",
		"res://assets/plants/imported/eggplant/stage_2.png",
		"res://assets/plants/imported/eggplant/stage_3.png",
		"res://assets/plants/imported/eggplant/stage_4.png",
	],
	"tomato": [
		"res://assets/plants/imported/tomato/stage_1.png",
		"res://assets/plants/imported/tomato/stage_2.png",
		"res://assets/plants/imported/tomato/stage_3.png",
		"res://assets/plants/imported/tomato/stage_4.png",
	],
	"strawberry": [
		"res://assets/plants/imported/strawberry/stage_1.png",
		"res://assets/plants/imported/strawberry/stage_2.png",
		"res://assets/plants/imported/strawberry/stage_3.png",
		"res://assets/plants/imported/strawberry/stage_4.png",
	],
	"corn": [
		"res://assets/plants/imported/corn/stage_1.png",
		"res://assets/plants/imported/corn/stage_2.png",
		"res://assets/plants/imported/corn/stage_3.png",
		"res://assets/plants/imported/corn/stage_4.png",
	],
	"sunflower": [
		"res://assets/plants/imported/sunflower/stage_1.png",
		"res://assets/plants/imported/sunflower/stage_2.png",
		"res://assets/plants/imported/sunflower/stage_3.png",
		"res://assets/plants/imported/sunflower/stage_4.png",
	],
	"pumpkin": [
		"res://assets/plants/imported/pumpkin/stage_1.png",
		"res://assets/plants/imported/pumpkin/stage_2.png",
		"res://assets/plants/imported/pumpkin/stage_3.png",
		"res://assets/plants/imported/pumpkin/stage_4.png",
	],
	"watermelon": [
		"res://assets/plants/imported/watermelon/stage_1.png",
		"res://assets/plants/imported/watermelon/stage_2.png",
		"res://assets/plants/imported/watermelon/stage_3.png",
		"res://assets/plants/imported/watermelon/stage_4.png",
	],
}

const WHEAT_STAGE_TEXTURES: Array[Texture2D] = [
	preload("res://assets/plants/user/crop_sheet_clean.png"),
	preload("res://assets/plants/user/crop_sheet_clean.png"),
	preload("res://assets/plants/user/crop_sheet_clean.png"),
	preload("res://assets/plants/user/crop_sheet_clean.png"),
]

const CORN_STAGE_TEXTURES: Array[Texture2D] = [
	preload("res://assets/plants/crops_extracted/corn_stage_1.png"),
	preload("res://assets/plants/crops_extracted/corn_stage_2.png"),
	preload("res://assets/plants/crops_extracted/corn_stage_3.png"),
	preload("res://assets/plants/crops_extracted/corn_stage_4.png"),
]

const CARROT_STAGE_TEXTURES: Array[Texture2D] = [
	preload("res://assets/plants/crops_extracted/carrot_stage_1.png"),
	preload("res://assets/plants/crops_extracted/carrot_stage_2.png"),
	preload("res://assets/plants/crops_extracted/carrot_stage_3.png"),
	preload("res://assets/plants/crops_extracted/carrot_stage_4.png"),
]

const PUMPKIN_STAGE_TEXTURES: Array[Texture2D] = [
	preload("res://assets/plants/crops_extracted/pumpkin_stage_1.png"),
	preload("res://assets/plants/crops_extracted/pumpkin_stage_2.png"),
	preload("res://assets/plants/crops_extracted/pumpkin_stage_3.png"),
	preload("res://assets/plants/crops_extracted/pumpkin_stage_4.png"),
]

static var _imported_texture_cache: Dictionary = {}

static func get_stage_texture(crop_key: String, stage: int) -> Texture2D:
	var paths: Array = IMPORTED_CROP_STAGE_PATHS.get(crop_key, [])
	if paths.is_empty():
		return null
	var index := clampi(stage, 0, paths.size() - 1)
	var path: String = str(paths[index])
	if not _imported_texture_cache.has(path):
		_imported_texture_cache[path] = load(path)
	return _imported_texture_cache[path] as Texture2D

static func get_wheat_stage_texture(stage: int) -> Texture2D:
	var index := clampi(stage, 0, WHEAT_STAGE_TEXTURES.size() - 1)
	return WHEAT_STAGE_TEXTURES[index]

static func get_corn_stage_texture(stage: int) -> Texture2D:
	var index := clampi(stage, 0, CORN_STAGE_TEXTURES.size() - 1)
	return CORN_STAGE_TEXTURES[index]

static func get_carrot_stage_texture(stage: int) -> Texture2D:
	var index := clampi(stage, 0, CARROT_STAGE_TEXTURES.size() - 1)
	return CARROT_STAGE_TEXTURES[index]

static func get_pumpkin_stage_texture(stage: int) -> Texture2D:
	var index := clampi(stage, 0, PUMPKIN_STAGE_TEXTURES.size() - 1)
	return PUMPKIN_STAGE_TEXTURES[index]
