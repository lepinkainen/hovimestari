# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hovimestari ("Butler" in Finnish) is a Go-based personal AI butler assistant inspired by Geoffrey Litt's Stevens assistant. It stores "memories" in SQLite, imports data from multiple sources (calendars, weather, manual input), and generates personalized daily briefs using the Google Gemini API. The project emphasizes simplicity with a pure Go implementation for easy cross-compilation.

See `PROJECT.md` for scope and direction.

## Development Workflow

Uses [Task](https://taskfile.dev/) rather than Make — see `Taskfile.yml` for the full command list. `task build` depends on lint and test passing.

**Always run `task build` (lint + test) before committing.**

Direct CLI usage: `./build/hovimestari <command> --config=/path/to/config.json --log-level=debug`

## Architecture

**CLI framework**: `alecthomas/kong`, not Cobra. Global flags: `--config`, `--log-level`.

**Core data flow**:

1. **Import**: importers fetch data → format as memories → store in SQLite
2. **Brief generation**: `internal/brief/brief.go` queries relevant memories → combines with prompts → sends to LLM → writes to every configured output

**Design principles**:

- **Two-table design**: the `memories` table holds general memories (weather, manual entries) with a hierarchical `source` field (e.g. `weather:helsinki`, `manual`); `calendar_events` holds structured calendar data with real datetime columns. Calendar events are the one structured exception — resist adding more tables.
- **Pure Go**: `modernc.org/sqlite`, no CGO, so `GOOS=linux GOARCH=amd64 go build` cross-compiles without Docker.
- **XDG compliance**: config lives in `~/.config/hovimestari/`.
- **Extensible I/O**: the output system writes to multiple destinations at once.

## Configuration

**Files**: `config.json`, `prompts.json`, `memories.db` (SQLite). Fields are documented in `config.example.json`; the struct lives in `internal/config/viper.go`.

**Resolution order**:

1. `--config` flag path
2. `$XDG_CONFIG_HOME/hovimestari/` (usually `~/.config/hovimestari/`)
3. Directory containing the executable

## Development Guidelines

**Code style**: follow `.clinerules/go-codestyle.md`. In particular:

- Wrap errors as `fmt.Errorf("failed to X: %w", err)`
- Prefer the standard library; `alecthomas/kong` for CLI, `spf13/viper` for config, `modernc.org/sqlite` for SQLite (CGO-free)
- `slog` for logging, `fmt.Printf` for interactive output

**Testing**: tests must stay deterministic — no network, no live database, no LLM calls. Focus on deterministic functions (URL conversion, formatting, parsing).

**Adding features**:

- **New importer**: new package under `internal/importer/`, mirroring the calendar importer's shape
- **New command**: new file in `cmd/hovimestari/commands/`, following the Kong pattern
- **New config field**: update the `Config` struct in `internal/config/viper.go` and `config.example.json`
