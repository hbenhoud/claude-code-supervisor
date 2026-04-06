# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Claude Code Supervisor — a companion dashboard that visualizes Claude Code sessions in real time via animated bot avatars. Read-only sidecar using Claude Code hooks.

## Architecture

- **Backend:** Go (stdlib net/http + gorilla/websocket + SQLite via mattn/go-sqlite3)
- **Frontend:** React 19 + TypeScript + Vite (in `web/`)
- **Data flow:** Claude Code hooks → POST /api/events → normalizer → SQLite → WebSocket → React dashboard

## Build & Run

```bash
# Dev mode (Go backend + Vite frontend with hot reload)
make dev

# Build single binary with embedded frontend
make build

# Run tests (exclude web/node_modules from Go)
go test $(go list ./... | grep -v node_modules)

# Frontend lint
npm --prefix web run lint
```

## Project Structure

```
cmd/supervisor/         — Main server entry point
cmd/supervisor-init/    — Hook installer CLI
internal/hooks/         — Hook event types + settings.json installer
internal/normalizer/    — Raw hook events → canonical SupervisorEvent
internal/store/         — SQLite persistence (sessions, events, scores)
internal/api/           — HTTP ingest + REST + WebSocket hub
internal/benchmark/     — Scoring engine + anti-pattern detection (TODO)
web/                    — React frontend (Vite)
```

## Key Conventions

- Backend is Go — never use Node.js for backend
- The Supervisor is a read-only observer — it never spawns or controls Claude Code
- Events come from Claude Code hooks, not stream-json
- Post-session enrichment reads cost/tokens from ~/.claude/ logs
- All metrics must be calculable from hook data or post-session enrichment
