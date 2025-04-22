// Package config provides configuration management for the Hovimestari application.
//
// This file is kept for backward compatibility. All configuration functionality
// has been moved to viper.go, which uses the Spf13/Viper library for more robust
// configuration management with features like:
// - Multiple search paths (XDG config directory, executable directory)
// - Automatic environment variable binding
// - Support for different configuration formats
//
// Please use the functions in viper.go (InitViper, GetConfig, LoadPrompts) for
// all configuration-related operations.
package config
