package main

import (
	"fmt"
	"os"

	"github.com/shrike/hovimestari/cmd/hovimestari/commands"
	"github.com/shrike/hovimestari/internal/config"
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
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to the configuration file")

	// Set up a PersistentPreRun function to initialize Viper
	// after the flags have been parsed
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize Viper with the config file path from the flag
		if err := config.InitViper(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
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
		fmt.Println(err)
		os.Exit(1)
	}
}
