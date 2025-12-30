# Repository Guidelines

## Project Structure & Module Organization
- Entry point lives in `cmd/hovimestari`, wiring the CLI with Kong.
- Domain logic sits under `internal/*`: `brief` assembles daily briefs, `importer` handles calendar/weather ingestion, `store` wraps SQLite access, and `output` manages delivery channels.
- Shared prompt artifacts reside in `llm-shared/`; background documentation sits in `docs/`.
- Tests stay alongside sources as `_test.go` files; builds land in `build/`, and `config.example.json` documents configurable fields.

## Build, Test, and Development Commands
- `task build` runs lint and tests before emitting `build/hovimestari`.
- `task run -- generate-brief` rebuilds then invokes the binary; append any CLI subcommand after `--`.
- `task test` wraps `go test ./...`; add flags like `-race` or `-run` when needed.
- `task lint` requires `golangci-lint` in `PATH`; install via `brew install golangci-lint` or the official script.
- For quick iteration use `go build ./cmd/hovimestari` or `go test ./internal/...` directly.

## Coding Style & Naming Conventions
- Always format with `gofmt` (or `goimports`); Task targets assume formatted code.
- Keep package names aligned with directory names (e.g., `internal/weather`, `internal/output`).
- Exported symbols use CamelCase plus short doc comments; file-level helpers stay lowerCamelCase.
- Configuration structs belong in `internal/config`; JSON keys mirror `config.example.json` using lower_case.

## Testing Guidelines
- Use Go's `testing` package with table-driven cases and `_test.go` suffixes.
- Stub external services by faking interfaces in `internal/llm` or `internal/importer`; avoid live API calls.
- Run `task test` (or `go test ./...`) before pushing; for coverage snapshots, run `go test -cover ./...`.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`feat:`, `fix:`, `chore:`) as in `git log`; keep subjects imperative under ~72 chars.
- Group related changes per commit and document breaking changes in the body if applicable.
- PRs should summarize behavior changes, list validation commands, and link issues (`Fixes #12`).
- Attach CLI transcripts or config snippets when altering user workflows or outputs.

## Configuration & Secrets
- Use `config.example.json` as the starting point; do not commit populated `config.json`, `.env`, or `memories.db`.
- Secrets load from `.env`, `$HOME/.hovimestari.env`, or environment variables before running `task` targets.
- Document new configuration flags in `docs/04_configuration.md` and update sample values in `config.example.json`.
