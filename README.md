# Claude Code Supervisor

> A companion dashboard that visualizes Claude Code sessions in real time.
> Watch your AI agents work as animated bots on a live mission control floor.

<!-- TODO: screenshot/gif -->

## What it does

- **Real-time visualization** of Claude Code tool calls, sub-agents, and file operations
- **Animated bot avatars** — each agent is a character with a name and visual personality
- **Post-session benchmarking** — execution score, error rate, retry analysis, anti-pattern detection
- **Session history** with full replay
- **Runs alongside your terminal** — never replaces Claude Code, never interferes

## Quick start

```bash
# Install hooks (one-time setup)
supervisor-init

# Start the dashboard
supervisor

# Use Claude Code in your terminal as usual
claude
# → Open http://localhost:3000 to watch
```

## How it works

Claude Code Supervisor uses Claude Code's **hook system** to observe tool calls without interfering with your terminal session. Hooks fire on every tool call and send events to the Supervisor backend via HTTP.

```
Terminal (you)            Supervisor (background)         Browser
    │                           │                           │
    │── use Claude Code ──→     │                           │
    │   hooks fire ──────→  POST /api/events                │
    │                           │── normalize ──→ SQLite     │
    │                           │── broadcast ──→ WebSocket ──→ Dashboard
    │                           │                           │
    │  (terminal unchanged)     │  (localhost:3001)         │  (localhost:3000)
```

**Key principle:** The Supervisor is a **read-only observer**. It never spawns, controls, or interferes with Claude Code. If the Supervisor is down, hooks fail silently — your terminal is unaffected.

## Dashboard views

### The Floor (live)

<!-- TODO: screenshot -->

A spatial workspace where each agent is an animated bot at a workstation. The root agent sits at the center; sub-agents appear at satellite stations connected by glowing data conduits. Bots animate based on what they're doing: typing for edits, scanning for reads, pulling a lever for Bash commands.

**4 bot variants:**
- **Root bot** — larger, primary color, the main agent
- **Explorer bot** — blue, magnifier motif, for Explore sub-agents
- **Planner bot** — orange, blueprint motif, for Plan sub-agents
- **Worker bot** — green, wrench motif, for general-purpose sub-agents

### Session Report (post-session)

<!-- TODO: screenshot -->

After a session ends, the Supervisor computes an **Execution Score** (S/A/B/C/D) from 4 metrics, detects anti-patterns, and compares with similar past sessions.

### Session List

<!-- TODO: screenshot -->

Browse active and past sessions. Active sessions show a live indicator. Click any session to watch live or replay.

## Scoring metrics

All metrics are computed **post-session**. Each has a tooltip explaining what it measures, how it's calculated, and what good vs bad looks like.

| Metric | Label | What it measures | Weight |
|--------|-------|-----------------|--------|
| Error rate | "Error rate" | % of tool calls that resulted in an error (Grep/Glob empty results excluded) | 30% |
| Retry density | "Repeat attempts" | Average retries per unique (tool, file) pair | 30% |
| Cost per change | "Cost per change" | Session cost ÷ files modified (from post-session log enrichment) | 20% |
| Session duration | "Time spent" | Duration percentile vs similar past sessions | 20% |

**Descriptive stats** (not scored) are also shown: total actions, files touched, helpers spawned, read vs write ratio, tool breakdown, total cost, tokens.

## Anti-patterns detected

Each anti-pattern card shows what happened, links to the offending events on the timeline, and suggests what to do.

| Pattern | What it detects |
|---------|----------------|
| **Struggled with edits** | Same file edited 3+ times in a row |
| **Search frenzy** | 5+ Grep/Glob calls in a row with different patterns |
| **Full file read** | Reading 500+ lines without offset/limit |
| **Commands failing** | 3+ consecutive Bash commands with non-zero exit |
| **Stalled** | 60s+ gap between tool calls |

## Configuration

```bash
# Install hooks into ~/.claude/settings.json
supervisor-init

# Remove hooks cleanly (preserves your other hooks)
supervisor-init --uninstall

# Start dashboard on custom ports
supervisor --port 4000 --api-port 4001
```

The Supervisor stores data in `~/.claude-supervisor/data.db` (SQLite).

## Architecture

| Layer | Technology |
|-------|-----------|
| Backend | Go (stdlib `net/http` + gorilla/websocket + SQLite) |
| Frontend | React 19 + TypeScript + Zustand + SVG + Framer Motion |
| Storage | SQLite with WAL mode (`~/.claude-supervisor/data.db`) |
| Build | Single binary with embedded frontend (`go:embed`) |

```
cmd/
  supervisor/          → Main server (API + static files)
  supervisor-init/     → Hook installer CLI
internal/
  hooks/               → Hook event types + installer
  normalizer/          → Raw events → canonical format
  store/               → SQLite persistence
  benchmark/           → Scoring engine + anti-pattern detection
  api/                 → HTTP + WebSocket handlers
web/                   → React frontend (Vite)
```

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- npm

### Commands

```bash
# Dev mode (backend + frontend with hot reload)
make dev

# Build single binary with embedded frontend
make build

# Run all tests
make test

# Lint
make lint
```

## Roadmap

- [ ] **Phase 1: "Bots come alive"** — Live dashboard with bot avatars, session list, replay
- [ ] **Phase 2: "Score & Express"** — Benchmark engine, rich animations, user modes, session comparison
- [ ] **Phase 3: "Control & Optimize"** — Skill tracker, trend dashboard, multi-session monitoring

## License

MIT
