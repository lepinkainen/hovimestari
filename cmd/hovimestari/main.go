package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/logging"

	// Import SQLite driver
	_ "modernc.org/sqlite"
)

// Version is set at build time via -ldflags
var Version string

// VersionCmd defines the version command for Kong
type VersionCmd struct{}

// Run executes the version command
func (cmd *VersionCmd) Run() error {
	if Version != "" {
		fmt.Printf("%s\n", Version)
	} else {
		fmt.Println("unknown (built without version information)")
	}
	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("hovimestari"),
		kong.Description("Hovimestari - A personal AI butler assistant"),
		kong.UsageOnError(),
	)

	// Initialize config and logging before command execution
	// Skip initialization for version command as it doesn't need config
	if ctx.Command() != "version" {
		if err := initializeApp(cli.Config, cli.LogLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute the selected command
	err := ctx.Run()
	if err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

// initializeApp initializes configuration and logging
func initializeApp(configPath, logLevel string) error {
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
	if logLevel != "debug" { // Kong has default value "debug"
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
