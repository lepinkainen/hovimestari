# Hovimestari Code Improvement Plan

This document outlines potential areas for code improvement identified during a scan of the project codebase. These focus on enhancing code quality, maintainability, and efficiency without introducing new features.

## Medium Impact / Code Quality & Efficiency

1.  **Optimize Weather Data Retrieval:** In `importer/weather/weather.go`, the `GetLatestForecasts` and `DetectForecastChanges` functions fetch all memories and filter/process in Go. Add specific queries to `store.Store` (e.g., `GetLatestForecastsBySourceAndDate`, `GetForecastHistoryForDate`) to perform filtering/aggregation directly in the database for better performance, especially as memory count grows.
2.  **Refactor `llm.BuildBriefPrompt`:** Break down the prompt assembly logic. Consider using a struct instead of `map[string]string` for `userInfo` for better type safety.
3.  **Improve Weather Error Handling:** In `brief.BuildBriefContext`, handle errors from `weatherimporter.GetLatestForecasts`, `weatherimporter.DetectForecastChanges`, and `weather.GetCurrentDayHourlyForecast` more formally instead of just printing warnings. Decide if these errors should halt brief generation or if fallback values are acceptable.

## High Impact / New Features

4.  **Google Calendar Support (via API):** Implement direct integration with the Google Calendar API instead of relying on WebCal/iCal exports.
    - Add OAuth2 authentication flow for Google API access.
    - Create a new importer in `internal/importer/gcalendar/`.
    - Add configuration options for Google Calendar accounts.
    - Update the README.md with the new functionality.
    - This would allow direct access to calendars without requiring WebCal/iCal export.

## Medium Impact / Flexibility Enhancement

5.  **Support for Other LLMs:** Add support for multiple LLM providers beyond Google Gemini.
    - Refactor the LLM interface to be provider-agnostic.
    - Implement adapters for multiple LLM providers, prioritizing those with official Go libraries:
      - OpenAI (github.com/sashabaranov/go-openai)
      - Anthropic Claude (github.com/anthropics/anthropic-sdk-go)
      - Ollama (github.com/ollama/ollama/api) implementation plan: `docs/llm-ollama.md`
      - Mistral AI (github.com/mistralai/mistralai-go)
      - Cohere (github.com/cohere-ai/cohere-go)
    - Add configuration options to select the LLM provider.
    - Implement prompt template adjustments for different LLM providers.
    - Update documentation with supported LLMs and configuration instructions.
