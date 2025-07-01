# Project: Hovimestari - Personal AI Butler Assistant

For detailed project information, refer to the documentation files in the `docs/` directory:

- `docs/01_overview.md`: Project overview and technology stack
- `docs/02_architecture.md`: Directory structure, key files, and data flow
- `docs/03_core_concepts.md`: Core concepts (memories, importers, briefs)
- `docs/04_configuration.md`: Configuration and configuration file examples
- `docs/05_database.md`: Database schema
- `docs/06_llm.md`: LLM integration, providers, and prompt structure
- `docs/07_cli.md`: CLI commands

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

1. Create a new command file in `cmd/hovimestari/commands/` (e.g., `new_command.go`).
2. Implement the command using the Cobra framework, following the pattern of existing commands.
3. Add the command to the root command in the `main.go` file.
4. Update the README.md with the new command.

Example of a new command file structure:

```go
package commands

import (
	"fmt"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/spf13/cobra"
)

// NewCommandCmd returns the new command
func NewCommandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-command",
		Short: "Short description",
		Long:  `Longer description of the command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the configuration
			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get configuration: %w", err)
			}

			// Implement command functionality
			return runNewCommand(cmd.Context())
		},
	}

	// Add flags if needed
	// cmd.Flags().StringVar(&flagVar, "flag-name", "", "Flag description")

	return cmd
}

// runNewCommand implements the command functionality
func runNewCommand(ctx context.Context) error {
	// Implement the command logic here
	return nil
}
```

### Working with the Configuration System

The application uses Spf13/Viper for configuration management. When working with configuration:

1. Use `config.GetConfig()` to retrieve the current configuration.
2. Access configuration values through the returned struct.
3. For new configuration options:
   - Add the field to the `Config` struct in `internal/config/viper.go`
   - Add default values in the `InitViper` function if needed
   - Add environment variable binding if appropriate
   - Update validation functions if necessary
   - Update the example configuration file

### Ollama Integration

When working with Ollama LLM integration:

1. The `docs/llm-ollama.md` file contains comprehensive documentation about Ollama integration.
2. Configuration should support both Gemini and Ollama as LLM providers.
3. The LLM interface should be provider-agnostic, allowing seamless switching between providers.
4. Ensure prompts work well with both Gemini and Ollama models.
