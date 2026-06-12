# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go + Gin backend for a QQ Farm 2.5D game. The frontend is a separate Godot 4.6 GDScript project (located in a sibling directory "新建游戏项目"). The backend serves as the authoritative server for all game state — the client is a thin renderer.

## Build & Run

```bash
# Run the server (listens on :8080)
go run main.go

# Build binary
go build -o server .

# Tidy dependencies
go mod tidy
```

No tests, linter, or CI are configured yet.

## Tech Stack

- **Framework**: Gin v1.9.1
- **ORM**: GORM v1.25 + PostgreSQL driver
- **Database**: PostgreSQL
- **Cache**: Redis (go-redis v9)
- **WebSocket**: gorilla/websocket (planned, phase 4)
- **Auth**: JWT / golang-jwt (planned, phase 2)

## Architecture

Standard layered Go web service layout:

```
main.go              — entry point, wires up router + DB + Redis
config/              — env/config loading
models/              — GORM models (User, FarmPlot, Inventory, CropDef)
handlers/            — HTTP request handlers (thin, delegate to services)
services/            — business logic (farm actions, auth, growth tick)
middleware/          — JWT auth, CORS, request logging
routes/              — route registration, grouped by domain
ws/                  — WebSocket hub + connection manager
dto/                 — request/response structs with validation tags
```

### Communication Pattern

- **REST**: registration, login, shop transactions, inventory queries, land reclaim
- **WebSocket**: real-time game state sync, growth ticks pushed to client, friend notifications (steal crops)

### Server-Authoritative Model

All game logic runs on the server. The client sends action intents (plant, water, harvest, etc.) and the server validates, mutates state, and pushes the resulting state back. Crop growth is computed server-side based on timestamps — the client interpolates for smooth visuals.

## Key Conventions

- All API responses use `{"code": int, "message": string, "data": object}` envelope
- JWT token passed via `Authorization: Bearer <token>` header
- Database auto-migration via GORM at startup (no separate migration tool)
- Config loaded from environment variables or `.env` file via `os.Getenv`
- Gold/experience math follows the client's existing formulas (reclaim cost = 60 + 35*plotIndex, level scaling = 1.5x)
