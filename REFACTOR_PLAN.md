# 重构计划：确立单一权威 + 拆分客户端上帝对象

## 背景与目标

QQ 农场项目由两部分组成：
- 后端 `go-gin-demo`（Go + Gin + GORM + PostgreSQL），`services/farm_engine.go` 已实现完整游戏规则（生长、事件、产量、升级）。
- 前端 `godot-game-demo`（Godot 4.6 / GDScript），`farm.gd` 单文件 ~2870 行，同时承担渲染、UI、输入、相机、HTTP、存档、离线补算。

文档声称"服务端权威，客户端是瘦渲染器"，但实际是**半套**：游戏规则在前后端各实现一份，存在可绕过的客户端权威路径、规则漂移、画面跳变，以及难以维护的单体脚本。

目标：**后端成为唯一权威，前端退化为渲染器 + 输入转发器**，并按职责拆分前端。

## 已确认的真实漏洞

1. **开垦地块完全客户端权威**：前端 `_try_reclaim_plot`（farm.gd:1337）只本地扣金币、本地解锁、本地存档，不走 `_send_action`；后端 `ExecuteAction` switch（farm_engine.go:281）无 `reclaim` case。改内存即可白嫖土地。
2. **`/farm/save` 让客户端覆盖整个存档**：`SaveFarm`（services/farm.go:16）无条件信任客户端传来的 gold/level/全部 plots/inventory，绕过一切服务端校验。前端每次操作后都调它。
3. **前端本地规则在登录态是死代码**，但仍在 `_process()`/`_check_events()` 用 `randf()` 运行，每 30 秒被 `_cloud_load` 覆盖，造成画面跳变。
4. **规则双写漂移**：阶段阈值（前端多一档 0.72）、事件 stageMult（后端 `stageMultBug[0]=0`）、浇水保护时长分档，前后端不一致。

---

## 阶段 0：堵权威漏洞（最高优先，独立可交付）

后端为主，改动集中、价值最高。

- 新增 `reclaim` action：`ExecuteAction` switch 加 `doReclaim`，复刻公式（费用 `60 + index*35`，等级 `index+1`），校验金币、等级、强制顺序解锁。
- 锁死 `/farm/save`：只接受非权威字段（`selected_seed`、`tool_mode`），忽略 gold/level/plots/inventory。
- 前端 `_try_reclaim_plot` 改为 `_send_action("reclaim", {plot_index})`，删除本地扣金币逻辑。

关键文件：`services/farm_engine.go:281`、`services/farm.go:16`、`routes/routes.go:33`、`dto/farm.go`、`farm.gd:1337`。

## 阶段 1：作物配置单一数据源

- 后端 `CROP_CONFIGS`/`FERTILIZER_COSTS`/阶段阈值暴露为 `/api/v1/farm/config`（或 login 时下发）。
- 前端启动拉取填充 `CROPS`，删除硬编码 `CROPS` 数组及本地规则函数（`_get_crop_stage_enum`/`_check_events`/`_apply_offline_growth`/`_calc_harvest_yield`/`_check_lv`/`_harvest_all`/`_shovel_all_crops`）。
- 统一阶段阈值：渲染只需 progress，不需要"逻辑阶段"。

## 阶段 2：前端只渲染服务端状态

- `_process()` 仅保留视觉进度插值（两次 `_apply_state` 之间线性推进 progress），不再产生事件、不改 inventory/gold。
- 删除 `randf()` 事件生成与离线补算（离线增长后端 `ProcessFarm` 已按 `LastProcessedAt` 算好）。
- 30 秒轮询保留为状态校正；因前端无竞争性本地状态，跳变消失。

## 阶段 3：修复 HTTP 并发竞态

单 `HTTPRequest` + 单 `_http_cb`（farm.gd:1245）会互相顶掉回调。

- 抽 `FarmApi`（autoload 或 Node）封装所有通信。
- 用请求队列串行发送，或 action/load/save 各开一个 `HTTPRequest`；回调带请求类型标识避免错配。

## 阶段 4：拆分 farm.gd 上帝对象

按职责拆分，每步可独立验证：
- `FarmApi.gd` — 后端通信（阶段 3 产物）
- `FarmRenderer.gd` — `_draw_world` 世界渲染
- `FarmHUD`（Control 子树）— 工具栏/HUD/确认框/overlay，**用 Control 节点替代手绘 + 手写 Rect2 命中**，消除 draw/input 坐标双写
- `CameraController.gd` — 鼠标/触屏/缩放
- `farm.gd` 瘦身为协调者：持有服务端 state，分发输入意图，通知渲染
- 删除 `_draw_input_debug_overlay` 和 `_debug_last_*`（farm.gd:2399）

## 阶段 5：统一存档/传输格式

- 删除前端本地二维数组存档路径（`_save_game`/`_restore_farm`），离线展示态也用 plots 结构。
- 消除 `_restore_farm` vs `_apply_cloud_data` 重复代码，及 `fert_stage_used` 双重 JSON 编码（farm.gd:1645、farm.gd:1769）。

---

## 执行顺序

阶段 0（堵漏洞）→ 阶段 1+2（消灭双写和跳变）→ 阶段 3 → 阶段 4（最大，纯前端，可分多次）→ 阶段 5 收尾。

## 验证方式

- 后端：`go build ./...` + `go run main.go`；curl 打 `/api/v1/farm/action`（plant/water/harvest/reclaim）核对 gold/plots；重点测 reclaim 的金币不足、等级不足、乱序解锁三种拒绝路径。
- 前端：Godot F5（或 godot-mcp `run_project` + `get_runtime_screenshot`），登录态走 种植→浇水→收获→开垦→卖出，确认无跳变、操作均经服务端。
- 项目无现成测试框架；阶段 0/1 后端规则建议补 Go 单元测试（`CalcHarvestYield`、`doReclaim` 校验、`ProcessFarm` 时间差）。
