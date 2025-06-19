# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hovimestari is a Go-based personal AI butler assistant that generates daily briefs. It stores "memories" in SQLite, imports calendar events and weather data, and uses Google Gemini API to generate personalized daily briefs in Finnish.

## Common Commands

Build and development:

- `task build` - Build the application for current OS/ARCH
- `task build-linux` - Cross-compile for Linux AMD64
- `task test` - Run all tests
- `task lint` - Run golangci-lint (requires golangci-lint installed)
- `task deps` - Tidy Go module dependencies

Application commands:

- `task import-calendar` - Import calendar events from WebCal URLs
- `task import-weather` - Import weather forecasts from MET Norway API
- `task generate-brief` - Generate and send daily brief
- `task add-memory CONTENT="text" RELEVANCE_DATE="2025-01-01" SOURCE="manual"` - Add memory manually
- `task init-config` - Initialize configuration (reads GEMINI_API_KEY, WEBCAL_URL from env)

The application uses Task runner (Taskfile.yml) instead of Make. All CLI commands also work directly: `./build/hovimestari import-calendar --config=/path/to/config.json --log-level=debug`

## Architecture

**Entry Point**: `cmd/hovimestari/main.go` - Uses Cobra CLI framework with global flags for config path and log level

**Core Components**:

- `internal/store/store.go` - SQLite database operations for single `memories` table
- `internal/config/viper.go` - Configuration management using Spf13/Viper with XDG directory support
- `internal/brief/brief.go` - Daily brief generation logic combining memories with LLM
- `internal/llm/gemini.go` - Google Gemini API client
- `internal/output/` - Multi-destination output system (CLI, Discord webhooks, Telegram bots)

**Importers**:

- `internal/importer/calendar/` - WebCal/iCalendar event importing with smart vs full_refresh modes
- `internal/importer/weather/` - MET Norway API weather forecast importing

**Key Design Patterns**:

- Single SQLite table (`memories`) stores all data with source prefixes for organization
- XDG Base Directory Specification for config file locations
- Modular output system supporting multiple destinations simultaneously
- Pure Go SQLite via modernc.org/sqlite (no CGO) for easy cross-compilation

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

## Development Notes

- Uses structured logging via `log/slog` with custom human-readable handler
- SQLite queries use prepared statements and proper error handling
- Calendar import supports both static calendars (smart updates) and dynamic ones (full refresh)
- Weather data formatted for Finnish users with metric units
- LLM prompts stored in `prompts.json` for easy modification
- Cross-platform build support without CGO dependencies
- Always run "task test" before committing to ensure deterministic behavior
- Always run "task lint" before committing to ensure code quality
- Use gofmt for formatting: `gofmt -w .`
