# CLI Commands

The application provides several CLI commands through the Cobra framework:

- **import-calendar**: Import events from configured WebCal URLs
- **import-weather**: Import weather forecasts for the configured location
- **generate-brief**: Generate a daily brief based on stored memories
  - `--days-ahead`: Number of days ahead to include in the brief (overrides config value)
- **add-memory**: Add a memory manually
- **init-config**: Initialize the configuration file
  - `--output-format`: Output format (cli, telegram)
- **list-models**: List available Gemini LLM models
- **show-brief-context**: Show the context that would be sent to the LLM

All commands support a global `--config` flag to specify a custom configuration file path.
