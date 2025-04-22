package main

import (
	"fmt"
	"os"

	"github.com/shrike/hovimestari/cmd/hovimestari/commands"
	"github.com/spf13/cobra"

	// Import SQLite driver
	_ "modernc.org/sqlite"
)

// ConfigPath is the path to the configuration file, exported for use by commands
var ConfigPath string

func main() {
	// Define the root command
	rootCmd := &cobra.Command{
		Use:   "hovimestari",
		Short: "Hovimestari - A personal AI butler assistant",
		Long:  `Hovimestari is a personal AI butler assistant that provides daily briefs and responds to queries.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.json", "Path to the configuration file")

	// Set up a PersistentPreRun function to set the ConfigPath in the commands package
	// after the flags have been parsed
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		commands.ConfigPath = ConfigPath
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
