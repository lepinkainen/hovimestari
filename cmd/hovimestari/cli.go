package main

import "github.com/lepinkainen/hovimestari/cmd/hovimestari/commands"

// CLI defines the main command structure for Kong CLI framework
type CLI struct {
	Config   string `kong:"help='Path to configuration file',short='c'"`
	LogLevel string `kong:"help='Log level (debug, info, warn, error)',default='debug'"`

	ImportCalendar     commands.ImportCalendarCmd     `kong:"cmd,help='Import calendar events from WebCal URLs'"`
	ImportWeather      commands.ImportWeatherCmd      `kong:"cmd,help='Import weather forecasts from MET Norway API'"`
	ImportWaterQuality commands.ImportWaterQualityCmd `kong:"cmd,help='Import water quality data for specific locations'"`
	GenerateBrief      commands.GenerateBriefCmd      `kong:"cmd,help='Generate and send daily brief'"`
	ShowBriefContext   commands.ShowBriefContextCmd   `kong:"cmd,help='Show context given to LLM without generating brief'"`
	AddMemory          commands.AddMemoryCmd          `kong:"cmd,help='Add memory manually to database'"`
	InitConfig         commands.InitConfigCmd         `kong:"cmd,help='Initialize configuration file'"`
	ListModels         commands.ListModelsCmd         `kong:"cmd,help='List available Gemini models'"`
	TUI                commands.TUICmd                `kong:"cmd,help='Start interactive terminal UI'"`
}