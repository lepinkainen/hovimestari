package xdg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// AppName is the application name used for XDG directories
	AppName = "hovimestari"
)

// GetConfigDir returns the config directory for the application.
// On macOS, it forces the use of $HOME/.config/hovimestari.
// On other systems, it follows the XDG Base Directory Specification:
// 1. If $XDG_CONFIG_HOME is set, use $XDG_CONFIG_HOME/hovimestari
// 2. Otherwise, use $HOME/.config/hovimestari
// The directory will be created if it doesn't exist.
func GetConfigDir() (string, error) {
	var appConfigDir string

	if runtime.GOOS == "darwin" {
		// Force ~/.config on macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		appConfigDir = filepath.Join(homeDir, ".config", AppName)
	} else {
		// Use standard XDG logic for other OSes
		configDir, err := os.UserConfigDir() // Respects XDG_CONFIG_HOME or defaults to ~/.config
		if err != nil {
			return "", fmt.Errorf("failed to get user config directory: %w", err)
		}
		appConfigDir = filepath.Join(configDir, AppName)
	}

	// Create the application-specific config directory
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory '%s': %w", appConfigDir, err)
	}

	return appConfigDir, nil
}

// GetExecutableDir returns the directory containing the executable
func GetExecutableDir() (string, error) {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Get the directory containing the executable
	return filepath.Dir(execPath), nil
}

// GetConfigPath returns the path to a configuration file
// It joins the config directory (determined by GetConfigDir) with the given filename
func GetConfigPath(filename string) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, filename), nil
}

// FindConfigFile looks for a configuration file in the following order:
// 1. The specified path (if not empty)
// 2. The config directory (determined by GetConfigDir)
//   - On macOS: $HOME/.config/hovimestari
//   - On other systems: XDG config directory
//
// 3. The executable directory
// It returns the path to the first file found, or an error if none is found
func FindConfigFile(filename, specifiedPath string) (string, error) {
	// Check the specified path first
	if specifiedPath != "" {
		if _, err := os.Stat(specifiedPath); err == nil {
			return specifiedPath, nil
		}
	}

	// Check the config directory (macOS: ~/.config/hovimestari, others: XDG config dir)
	configDir, err := GetConfigDir()
	if err == nil {
		path := filepath.Join(configDir, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check the executable directory
	exeDir, err := GetExecutableDir()
	if err == nil {
		path := filepath.Join(exeDir, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// If we get here, the file wasn't found
	configDirMsg := "$HOME/.config/hovimestari (macOS) or XDG default (other OS)"
	exeDirMsg, _ := GetExecutableDir() // Ignore error for message
	return "", fmt.Errorf("file '%s' not found in specified path, %s, or %s", filename, configDirMsg, exeDirMsg)
}
