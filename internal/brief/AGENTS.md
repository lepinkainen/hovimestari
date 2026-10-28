# internal/brief — agent notes

`BuildBriefContext` gathers memories, calendar events and derived facts; `GenerateDailyBrief` hands them to the LLM. Nothing here calls the model directly.

- **Any new `userInfo` key must also be rendered in `internal/llm/gemini.go` `BuildBriefPrompt`.** The map is not dumped wholesale; the prompt builder reads a fixed list of keys and formats each one. A key set here but not read there silently never reaches the model. Add a test in `gemini_test.go` asserting the line appears, as `TestBuildBriefPromptIncludesDayType` does.
- **The prompt refers to context fields by name** (`'Current Time'`, `'Day Type'`, `'HourlyForecastToday'`). If you rename a key or its rendered label, update `prompts.json` in the same change.
- **Source-prefix filtering lives in `getRelevantMemoryStrings`.** Weather memories are skipped there because weather is injected via `userInfo` instead; school lunch is restricted to today. Add new per-source rules in that function, not in the store queries.
- **Dates are computed in the configured timezone** (`cfg.Timezone`); keep `now` flowing through parameters so tests can pin it. `assembleUserInfo` and `dayType` take `now` explicitly for this reason.
- **Tests build a real `Generator` over a temp SQLite store** with `llm` set to nil (`newBriefGenerator` in `brief_test.go`). Test the context-building functions; never call `GenerateDailyBrief` in tests, it needs the network.
