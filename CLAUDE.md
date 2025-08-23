# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hovimestari ("Butler" in Finnish) is a Go-based personal AI butler assistant inspired by Geoffrey Litt's Stevens assistant. It stores "memories" in a single SQLite table, imports data from multiple sources (calendars, weather, manual input), and generates personalized daily briefs using Google Gemini API. The project emphasizes simplicity with a pure Go implementation for easy cross-compilation.

## Development Workflow

**Build System**: Uses [Task](https://taskfile.dev/) runner instead of Make (see `Taskfile.yml`). Build depends on lint and test passing.

**Critical Commands**:
- `task build` - Build for current OS/ARCH (runs lint + test first)
- `task build-linux` - Cross-compile for Linux AMD64 (CGO-free)
- `task test` - Run all tests (deterministic, no external deps)
- `task lint` - Run golangci-lint (required before commit)
- `task deps` - Tidy Go module dependencies

**Application Commands**:
- `task import-calendar` - Import WebCal/iCalendar events with smart/full_refresh modes
- `task import-weather` - Import MET Norway weather forecasts  
- `task import-water-quality` - Import water quality data for specific locations
- `task generate-brief` - Generate daily brief using LLM and current memories
- `task add-memory CONTENT="text" RELEVANCE_DATE="2025-01-01" SOURCE="manual"` - Add manual memory
- `task init-config` - Initialize config (reads GEMINI_API_KEY, WEBCAL_URL from env)

**Direct CLI Usage**: `./build/hovimestari <command> --config=/path/to/config.json --log-level=debug`

## Architecture

**CLI Framework**: Uses `alecthomas/kong` (not Cobra) for command parsing in `cmd/hovimestari/main.go`. Global flags: `--config`, `--log-level`

**Core Data Flow**:
1. **Import Phase**: Various importers fetch data → format as memories → store in SQLite
2. **Brief Generation**: `internal/brief/brief.go` queries relevant memories → combines with prompts → sends to LLM → outputs to multiple destinations

**Key Components**:

- `internal/store/store.go` - Single SQLite table (`memories`) with source-based organization
- `internal/config/viper.go` - Viper configuration with XDG Base Directory support  
- `internal/brief/brief.go` - Brief generation orchestrator combining memories + LLM
- `internal/llm/gemini.go` - Google Gemini API client (supports multiple models)
- `internal/logging/handler.go` - Custom slog handler for human-readable output
- `internal/output/` - Multi-destination system (CLI, Discord, Telegram)

**Importers Pattern**:
- `internal/importer/calendar/` - WebCal imports with smart (upsert) vs full_refresh (replace_all) strategies
- `internal/importer/weather/` - MET Norway API integration
- Commands in `cmd/hovimestari/commands/` for manual data entry

**Design Principles**:
- **Single Table Design**: All data in `memories` table with hierarchical `source` field (e.g., "calendar:work", "weather:helsinki")
- **Pure Go**: Uses `modernc.org/sqlite` (no CGO) for cross-compilation without Docker
- **XDG Compliance**: Config files follow standard (`~/.config/hovimestari/`)
- **Extensible I/O**: Output system supports multiple simultaneous destinations

## Configuration

**Files**: `config.json`, `prompts.json`, `memories.db` (SQLite)

**Config Resolution Order**:

1. `--config` flag path
2. `$XDG_CONFIG_HOME/hovimestari/` (usually `~/.config/hovimestari/`)
3. Directory containing executable

**Key Config Fields**:

- `gemini_api_key`, `gemini_model` - LLM configuration
- `calendars[]` with `update_mode: "smart"|"full_refresh"` - Calendar import strategy
- `outputs.enable_cli`, `outputs.discord_webhook_urls[]`, `outputs.telegram_bots[]` - Multi-destination output
- `family[]` with optional birthdays - Birthday tracking in briefs

## Memory System

All data is stored as "memories" in SQLite with:

- `content` - Formatted text (e.g., "Calendar Event: Meeting from 2025-01-01 14:00 to 15:00")
- `source` - Hierarchical source identifier (e.g., "calendar:work", "weather:helsinki", "manual")
- `relevance_date` - When memory is relevant (used for brief filtering)
- `uid` - Optional unique identifier (prevents calendar event duplicates)

## Testing

Tests exist for deterministic functions in `*_test.go` files:

- Calendar URL conversion and event formatting
- Weather forecast formatting
- Output system behavior

Run with `task test` or `go test ./...`. Tests avoid external dependencies (network, database, LLM calls).

## Development Guidelines

**Code Style**: Follow `.clinerules/go-codestyle.md` conventions:
- Use `fmt.Errorf("failed to X: %w", err)` for error wrapping
- Prefer standard library, use `alecthomas/kong` for CLI, `spf13/viper` for config
- Use `modernc.org/sqlite` for SQLite (CGO-free)
- Use `slog` for logging, `fmt.Printf` for interactive output

**Testing Strategy**: 
- Tests avoid external dependencies (network, database, LLM calls)
- Focus on deterministic functions: URL conversion, data formatting, parsing
- Examples: `calendar_test.go` (event formatting), `weather_test.go` (forecast formatting)
- Run `task test` (includes in build pipeline)

**Adding New Features**:
- **New Importer**: Create package in `internal/importer/`, implement similar interface to calendar importer
- **New Command**: Add file in `cmd/hovimestari/commands/`, follow Kong CLI pattern
- **New Config**: Update `Config` struct in `internal/config/viper.go`, add to `config.example.json`

**Memory Storage Pattern**:
```go
// All data stored as memories with structured content
content := fmt.Sprintf("Calendar Event: %s from %s to %s", summary, start, end)
source := "calendar:" + calendarName  // Hierarchical source naming
uid := event.UID  // Optional unique identifier for deduplication
```

**Cross-Compilation**: Pure Go implementation enables simple `GOOS=linux GOARCH=amd64 go build` without Docker or CGO

**Commit Requirements**: Always run `task build` (includes lint + test) before commits
