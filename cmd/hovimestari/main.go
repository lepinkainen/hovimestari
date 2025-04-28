package main

import (
	"log/slog"
	"os"

	"github.com/shrike/hovimestari/cmd/hovimestari/commands"
	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/logging"
	"github.com/spf13/cobra"

	// Import SQLite driver
	_ "modernc.org/sqlite"
)

func main() {
	// Initialize the default logger with our custom human-readable handler
	logger := slog.New(logging.NewHumanReadableHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Define the root command
	rootCmd := &cobra.Command{
		Use:   "hovimestari",
		Short: "Hovimestari - A personal AI butler assistant",
		Long:  `Hovimestari is a personal AI butler assistant that provides daily briefs and responds to queries.`,
	}

	// Add global flags
	var configPath string
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to the configuration file")

	// Set up a PersistentPreRun function to initialize Viper
	// after the flags have been parsed
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize Viper with the config file path from the flag
		if err := config.InitViper(configPath); err != nil {
			slog.Error("Error initializing configuration", "error", err)
			os.Exit(1)
		}
	}

	// Add commands
	rootCmd.AddCommand(commands.ImportCalendarCmd())
	rootCmd.AddCommand(commands.ImportWeatherCmd())
	rootCmd.AddCommand(commands.GenerateBriefCmd())
	rootCmd.AddCommand(commands.ShowBriefContextCmd())
	rootCmd.AddCommand(commands.AddMemoryCmd())
	rootCmd.AddCommand(commands.InitConfigCmd())
	rootCmd.AddCommand(commands.ListModelsCmd())

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}
