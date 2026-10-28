# internal/llm — agent notes

Thin wrapper over `google.golang.org/genai`. Keep it thin.

- **Prompts live in `prompts.json`, not in Go.** This package only substitutes placeholders (`%CONTEXT%`, `%NOTES%`, `%LANG%`, `%QUERY%`) and renders the context block. Tone and structure instructions belong in the JSON file so they can change without a rebuild.
- **`BuildBriefPrompt` renders an explicit list of `userInfo` keys.** When `internal/brief` adds a key, add a matching line here and a test. See the note in `internal/brief/AGENTS.md`.
- **Tests must not touch the network.** Build a `Client` literal with only `prompts` set and test the pure functions (`BuildBriefPrompt`, `cleanMarkdownWrapper`). Never call `NewClient`, `Generate` or `ListModels` in tests; there is no API key in CI.
- **SDK notes:** the client has no persistent connection, so `Close` is a deliberate no-op kept for callers. `ListModels` uses the SDK's paging iterator and must surface its error rather than stop silently. The old `github.com/google/generative-ai-go` is deprecated and does not compile against current Google modules; do not reintroduce it.
