package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/shrike/hovimestari/cmd/hovimestari/commands"
	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/logging"
	"github.com/spf13/cobra"

	// Import SQLite driver
	_ "modernc.org/sqlite"
)

func main() {
	// Define the root command
	rootCmd := &cobra.Command{
		Use:   "hovimestari",
		Short: "Hovimestari - A personal AI butler assistant",
		Long:  `Hovimestari is a personal AI butler assistant that provides daily briefs and responds to queries.`,
	}

	// Add global flags
	var configPath string
	var logLevel string
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to the configuration file")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error)")

	// Set up a PersistentPreRunE function to initialize the logger and Viper
	// after the flags have been parsed
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Initialize Viper with the config file path from the flag
		if err := config.InitViper(configPath); err != nil {
			// Use a basic logger for this error since the full logger isn't set up yet
			fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
			return err
		}

		// Get the configuration to check for log level
		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting configuration: %v\n", err)
			return err
		}

		// Determine the log level to use
		// Command line flag takes precedence over config file
		var logLevelToUse string
		if cmd.Flags().Changed("log-level") {
			logLevelToUse = logLevel
		} else if cfg.LogLevel != "" {
			logLevelToUse = cfg.LogLevel
		} else {
			logLevelToUse = "info" // Default if not specified in flag or config
		}

		// Set the log level based on the determined value
		var level slog.Level
		switch strings.ToLower(logLevelToUse) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}

		// Create and set the logger with the appropriate level
		opts := &slog.HandlerOptions{
			Level: level,
		}
		logger := slog.New(logging.NewHumanReadableHandler(os.Stderr, opts))
		slog.SetDefault(logger)

		// Log the selected log level
		slog.Debug("Logger initialized", "level", logLevelToUse)

		return nil
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
