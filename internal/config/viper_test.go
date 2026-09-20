package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lepinkainen/hovimestari/internal/xdg"
	"github.com/spf13/viper"
)

func isolateConfig(t *testing.T) string {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "HOVIMESTARI_") {
			t.Setenv(key, "")
		}
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "appdata"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	dir, err := xdg.GetConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func validConfigData() map[string]any {
	return map[string]any{
		"gemini_api_key": "test-key",
		"location_name":  "Helsinki",
		"latitude":       60.1699,
		"longitude":      24.9384,
		"timezone":       "Europe/Helsinki",
		"calendars":      []CalendarConfig{{Name: "Home", URL: "https://example.invalid/calendar.ics"}},
		"family":         []FamilyMember{{Name: "Test", Birthday: "2000-01-01"}},
	}
}

func writeConfigFile(t *testing.T, path string, data any) {
	t.Helper()
	contents, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestConfigDefaultsAndEnvironmentOverrides(t *testing.T) {
	for _, override := range []bool{false, true} {
		name := "defaults"
		if override {
			name = "environment overrides file"
		}
		t.Run(name, func(t *testing.T) {
			configDir := isolateConfig(t)
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			data := validConfigData()
			if override {
				data["gemini_model"] = "file-model"
				data["log_level"] = "info"
				data["db_path"] = "file.db"
				t.Setenv("HOVIMESTARI_GEMINI_API_KEY", "env-key")
				t.Setenv("HOVIMESTARI_GEMINI_MODEL", "env-model")
				t.Setenv("HOVIMESTARI_LOG_LEVEL", "warn")
				t.Setenv("HOVIMESTARI_DB_PATH", "env.db")
			}
			writeConfigFile(t, path, data)
			writeConfigFile(t, filepath.Join(dir, "prompts.json"), map[string][]string{"dailyBrief": {"Test prompt"}})
			if err := InitViper(path); err != nil {
				t.Fatal(err)
			}
			cfg, err := GetConfig()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.LocationName != "Helsinki" || cfg.Timezone != "Europe/Helsinki" || len(cfg.Calendars) != 1 || len(cfg.Family) != 1 {
				t.Fatalf("file configuration not loaded: %+v", cfg)
			}
			if cfg.DaysAhead != 2 || cfg.OutputLanguage != "Finnish" || !cfg.Outputs.EnableCLI {
				t.Errorf("incorrect defaults: %+v", cfg)
			}
			if cfg.PromptFilePath != filepath.Join(dir, "prompts.json") {
				t.Errorf("prompt path = %q", cfg.PromptFilePath)
			}
			wantKey, wantModel, wantLog, wantDB := "test-key", "gemini-2.5-flash", "info", filepath.Join(configDir, "memories.db")
			if override {
				wantKey, wantModel, wantLog, wantDB = "env-key", "env-model", "warn", filepath.Join(dir, "env.db")
			}
			if cfg.GeminiAPIKey != wantKey || cfg.GeminiModel != wantModel || cfg.LogLevel != wantLog || cfg.DBPath != wantDB {
				t.Errorf("incorrect resolved configuration: %+v", cfg)
			}
		})
	}
}

func TestConfigDiscoveryPrecedence(t *testing.T) {
	configDir := isolateConfig(t)
	exeDir, err := xdg.GetExecutableDir()
	if err != nil {
		t.Fatal(err)
	}
	exeConfig := filepath.Join(exeDir, "config.json")
	// The test executable lives in Go's temporary build directory. Never overwrite an existing file.
	f, err := os.OpenFile(exeConfig, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(exeConfig); err != nil {
			t.Error(err)
		}
	})
	writeConfigFile(t, exeConfig, map[string]string{"location_name": "executable"})
	xdgConfig := filepath.Join(configDir, "config.json")
	writeConfigFile(t, xdgConfig, map[string]string{"location_name": "xdg"})
	explicit := filepath.Join(t.TempDir(), "chosen.json")
	writeConfigFile(t, explicit, map[string]string{"location_name": "explicit"})
	for _, tc := range []struct{ name, flag, want string }{
		{"explicit", explicit, "explicit"},
		{"xdg", "", "xdg"},
		{"executable", "", "executable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			if tc.name == "executable" {
				if err := os.Remove(xdgConfig); err != nil {
					t.Fatal(err)
				}
			}
			if err := InitViper(tc.flag); err != nil {
				t.Fatal(err)
			}
			if got := viper.GetString("location_name"); got != tc.want {
				t.Errorf("location = %q, want %q", got, tc.want)
			}
		})
	}
}

// A missing explicit file is an error; a missing default file permits env-only use.
func TestInitViperMissingConfig(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		name := "optional default"
		if explicit {
			name = "explicit path"
		}
		t.Run(name, func(t *testing.T) {
			isolateConfig(t)
			path := ""
			if explicit {
				path = filepath.Join(t.TempDir(), "missing.json")
			}
			err := InitViper(path)
			if (err != nil) != explicit {
				t.Errorf("error = %v, want error: %t", err, explicit)
			}
		})
	}
}

func TestConfigPathResolution(t *testing.T) {
	for _, mode := range []string{"relative", "absolute", "fallback", "missing"} {
		t.Run(mode, func(t *testing.T) {
			configDir := isolateConfig(t)
			dir := t.TempDir()
			data := validConfigData()
			wantDB := filepath.Join(dir, "data", "brief.db")
			wantPrompt := filepath.Join(dir, "templates", "custom.json")
			data["db_path"], data["prompt_file_path"] = "data/brief.db", "templates/custom.json"
			if mode == "absolute" {
				data["db_path"], data["prompt_file_path"] = wantDB, wantPrompt
			}
			if mode == "fallback" {
				wantPrompt = filepath.Join(configDir, "prompts.json")
			}
			if mode != "missing" {
				writeConfigFile(t, wantPrompt, map[string][]string{"dailyBrief": {"Hello"}})
			}
			path := filepath.Join(dir, "config.json")
			writeConfigFile(t, path, data)
			if err := InitViper(path); err != nil {
				t.Fatal(err)
			}
			cfg, err := GetConfig()
			if mode == "missing" {
				if err == nil || !strings.Contains(err.Error(), "prompts file not found") {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.DBPath != wantDB || cfg.PromptFilePath != wantPrompt {
				t.Errorf("paths = (%q, %q), want (%q, %q)", cfg.DBPath, cfg.PromptFilePath, wantDB, wantPrompt)
			}
		})
	}
}

func TestGetConfigValidation(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		value     any
		want      string
	}{
		{"API key", "gemini_api_key", "", "gemini API key is required"},
		{"location", "location_name", "", "location_name is required"},
		{"latitude", "latitude", 91, "latitude must be between"},
		{"longitude", "longitude", -181, "longitude must be between"},
		{"missing timezone", "timezone", "", "timezone is required"},
		{"invalid timezone", "timezone", "invalid/zone", "invalid timezone"},
		{"no calendars", "calendars", []CalendarConfig{}, "at least one calendar"},
		{"calendar name", "calendars", []CalendarConfig{{URL: "https://example.invalid"}}, "missing a name"},
		{"calendar URL", "calendars", []CalendarConfig{{Name: "Home"}}, "missing a URL"},
		{"no family", "family", []FamilyMember{}, "at least one family member"},
		{"family name", "family", []FamilyMember{{}}, "missing a name"},
		{"birthday", "family", []FamilyMember{{Name: "Test", Birthday: "not-a-date"}}, "invalid birthday"},
		{"water location", "water_quality_locations", []WaterQualityLocation{{}}, "missing a name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolateConfig(t)
			writeConfigFile(t, filepath.Join(dir, "prompts.json"), map[string][]string{})
			data := validConfigData()
			data[tc.key] = tc.value
			path := filepath.Join(dir, "config.json")
			writeConfigFile(t, path, data)
			if err := InitViper(path); err != nil {
				t.Fatal(err)
			}
			_, err := GetConfig()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestConfigExplicitOutputs(t *testing.T) {
	dir := isolateConfig(t)
	data := validConfigData()
	data["outputs"] = OutputConfig{DiscordWebhookURLs: []string{"https://example.invalid/webhook"}}
	writeConfigFile(t, filepath.Join(dir, "config.json"), data)
	writeConfigFile(t, filepath.Join(dir, "prompts.json"), map[string][]string{})
	if err := InitViper(""); err != nil {
		t.Fatal(err)
	}
	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Outputs.EnableCLI || len(cfg.Outputs.DiscordWebhookURLs) != 1 {
		t.Errorf("explicit outputs changed: %+v", cfg.Outputs)
	}
}

func TestConfigKeySpellings(t *testing.T) {
	for _, keys := range []struct{ language, prompt string }{
		{"outputLanguage", "promptFilePath"},
		{"output_language", "prompt_file_path"},
	} {
		t.Run(keys.language, func(t *testing.T) {
			dir := isolateConfig(t)
			data := validConfigData()
			data[keys.language] = "English"
			data[keys.prompt] = "custom.json"
			writeConfigFile(t, filepath.Join(dir, "config.json"), data)
			writeConfigFile(t, filepath.Join(dir, "custom.json"), map[string][]string{})
			if err := InitViper(""); err != nil {
				t.Fatal(err)
			}
			cfg, err := GetConfig()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.OutputLanguage != "English" || cfg.PromptFilePath != filepath.Join(dir, "custom.json") {
				t.Errorf("key spelling ignored: language %q, prompt path %q", cfg.OutputLanguage, cfg.PromptFilePath)
			}
		})
	}
}

func TestLoadPrompts(t *testing.T) {
	for _, tc := range []struct{ name, content, wantErr string }{
		{"valid", `{"dailyBrief":["Hello","%NOTES%"],"userQuery":["%QUERY%"]}`, ""},
		{"default path", `{"dailyBrief":["Hello","%NOTES%"],"userQuery":["%QUERY%"]}`, ""},
		{"missing", "", "failed to open prompts file"},
		{"malformed", "{", "failed to decode prompts file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "prompts.json")
			if tc.content != "" {
				if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.name == "default path" {
				t.Chdir(dir)
				path = ""
			}
			got, err := LoadPrompts(path)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := map[string][]string{"dailyBrief": {"Hello", "%NOTES%"}, "userQuery": {"%QUERY%"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("prompts = %#v, want %#v", got, want)
			}
		})
	}
}
