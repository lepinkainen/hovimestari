# Improvement Ideas

- Fix the date-range predicate in `internal/store/store.go:319` so events that start before the window and end inside it are returned; swapping the final argument to `endDate` prevents multi-day events from being dropped.
- Inject context-aware HTTP clients with timeouts for external calls (e.g., `internal/importer/calendar/calendar.go:69`, `internal/weather/metno.go`, `internal/output/discord.go:47`) to avoid hangs and honor CLI cancellations.
- Deduplicate weather memories during imports (`internal/importer/weather/weather.go:35`) by checking for existing entries or pruning older ones to stop the database from growing without bound.
- Populate the `FutureWeather` and `WeatherChanges` fields expected in `internal/llm/gemini.go:140` by summarising upcoming forecasts in `assembleUserInfo` (`internal/brief/brief.go:122`), giving the LLM structured context.
- Use the configured timezone when computing relevance windows in `GenerateResponse` (`internal/brief/brief.go:311`) so ad-hoc answers match the household's locale, not the host system.
- Add store-layer tests with a temporary SQLite DB to cover calendar overlaps, weather dedupe behaviour, and guard the SQL in `internal/store/store.go` against regressions.
