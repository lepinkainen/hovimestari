# Project: Hovimestari - Personal AI Butler Assistant

For detailed project structure, technology stack, and core concepts, refer to `PROJECT_CONTEXT.md`.

## Language

- The name "Hovimestari" means "Butler" in Finnish.
- All briefs and responses are generated in a language the user can configure.
- Code comments and documentation are in English.
- All log output and CLI messages should be in English.

### LLM Prompts

- Store LLM prompts in English in configuration files (e.g., `prompts.json`).
- Use placeholders (e.g., `%LANG%`) for output language, which is specified in `config.json`.
- Use placeholders (e.g., `%NOTES%`) for dynamic content like memories.
- Keep prompts clear, concise, and focused on the specific task.

## Common Tasks

### Adding a New Importer

1. Create a new package under `internal/importer/`.
2. Implement the importer with a similar interface to the calendar importer.
3. Add a new command in `cmd/hovimestari/main.go`.
4. Update the README.md with the new functionality.

### Modifying the Database Schema

1. Update the `Initialize` function in `internal/store/store.go`.
2. Add any new methods needed to interact with the new schema.
3. Consider adding a migration mechanism if the change is not backward compatible.

### Adding a New Command

1. Create a new command function in `cmd/hovimestari/main.go`.
2. Add the command to the root command in the `main` function.
3. Implement the command's functionality.
4. Update the README.md with the new command.
