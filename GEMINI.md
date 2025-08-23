# Gemini Code Assistant Guide for Hovimestari

This guide provides essential information for AI code assistants working on the Hovimestari project.

## Project Overview

Hovimestari is a personal AI butler assistant written in Go. It gathers information from various sources (calendars, weather APIs, manual input), stores them as "memories" in a SQLite database, and uses a Large Language Model (LLM) to generate a personalized daily brief.

- **Inspiration**: [Geoffrey Litt's AI assistant](https://www.geoffreylitt.com/2025/04/12/how-i-made-a-useful-ai-assistant-with-one-sqlite-table-and-a-handful-of-cron-jobs)
- **Name**: "Hovimestari" is Finnish for "Butler".
- **Primary Language**: Briefs are generated in Finnish by default, but the codebase, comments, and logs are in English.

## Core Architecture & Key Files

The project follows a modular structure, separating concerns into distinct packages within the `internal/` directory.

- **Entry Point**: `cmd/hovimestari/main.go` initializes the [Cobra](https://github.com/spf13/cobra) CLI.
- **CLI Commands**: Each command is a separate file in `cmd/hovimestari/commands/`.
- **Configuration**: `internal/config/viper.go` manages configuration using [Viper](https://github.com/spf13/viper), supporting file-based (`config.json`), environment variables, and XDG standard directories (`~/.config/hovimestari/`).
- **Database**: `internal/store/store.go` handles all interactions with the SQLite database (`memories.db`). It uses `modernc.org/sqlite` for CGO-free compilation. All data is stored in a single `memories` table.
- **Brief Generation**: `internal/brief/brief.go` orchestrates the collection of memories and context to generate the final brief via the LLM.
- **LLM Interaction**: `internal/llm/gemini.go` contains the client for the Google Gemini API. Prompts are stored in `prompts.json`.
- **Importers**: Data sources are implemented as importers in `internal/importer/`. For example, `internal/importer/calendar/calendar.go` handles iCalendar/WebCal imports.
- **Output**: `internal/output/` contains a multi-destination system to send briefs to the CLI, Discord, and Telegram.

## Developer Workflow & Commands

This project uses [Task](https://taskfile.dev/) as a command runner instead of Make. All common development tasks are defined in `Taskfile.yml`.

- **Build the application**:

  ```bash
  task build
  ```

- **Run all tests**:

  ```bash
  task test
  ```

- **Run linter**:

  ```bash
  task lint
  ```

- **Tidy dependencies**:

  ```bash
  task deps
  ```

Application commands are executed via the compiled binary. For example:

```bash
# Import calendar events
./build/hovimestari import-calendar

# Generate the daily brief
./build/hovimestari generate-brief
```

## Project Conventions

- **Error Handling**: Use the `fmt.Errorf("...: %w", err)` pattern to wrap and add context to errors.
- **Logging**: The project uses the standard `log/slog` library with a custom human-readable handler in `internal/logging/handler.go`. Use this for all logging.
- **Dependencies**: Use the standard library where possible. For external dependencies, ensure they are added to `go.mod` and run `task deps`.
- **Configuration**: When adding new configuration options, update the `Config` struct in `internal/config/viper.go` and the `config.example.json` file.
- **Adding a Command**: To add a new CLI command, create a new file in `cmd/hovimestari/commands/` and add it to the root command in `cmd/hovimestari/main.go`. Follow the existing Cobra command structure.

## Testing

- Tests are located in `*_test.go` files alongside the code they test.
- Tests should be deterministic and not rely on external services (network, LLM APIs, or the database). Mock these dependencies where necessary.
- Run all tests using `task test`.
