package xdg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "appdata"))
	t.Setenv("XDG_CONFIG_HOME", "")
	return home
}

func TestGetConfigDir(t *testing.T) {
	for _, useXDG := range []bool{false, true} {
		name := "default"
		if useXDG {
			name = "XDG override"
		}
		t.Run(name, func(t *testing.T) {
			home := isolateHome(t)
			base := filepath.Join(home, ".config")
			if runtime.GOOS == "windows" {
				base = filepath.Join(home, "appdata")
			}
			if useXDG {
				override := filepath.Join(home, "custom")
				t.Setenv("XDG_CONFIG_HOME", override)
				if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
					base = override
				}
			}
			want := filepath.Join(base, AppName)
			for range 2 { // Directory creation is idempotent.
				got, err := GetConfigDir()
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("directory = %q, want %q", got, want)
				}
				info, err := os.Stat(got)
				if err != nil || !info.IsDir() {
					t.Fatalf("config directory was not created: %v", err)
				}
			}
			path, err := GetConfigPath("settings.json")
			if err != nil {
				t.Fatal(err)
			}
			if path != filepath.Join(want, "settings.json") {
				t.Errorf("config path = %q", path)
			}
		})
	}
}

func TestGetConfigDirErrors(t *testing.T) {
	t.Run("missing home", func(t *testing.T) {
		isolateHome(t)
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "")
		t.Setenv("APPDATA", "")
		if _, err := GetConfigDir(); err == nil {
			t.Error("expected missing home error")
		}
		if _, err := GetConfigPath("config.json"); err == nil {
			t.Error("GetConfigPath swallowed missing home error")
		}
	})
	t.Run("directory blocked by file", func(t *testing.T) {
		home := isolateHome(t)
		base := filepath.Join(home, ".config")
		if runtime.GOOS == "windows" {
			base = filepath.Join(home, "appdata")
		}
		if err := os.WriteFile(base, []byte("not a directory"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := GetConfigDir(); err == nil || !strings.Contains(err.Error(), "failed to create config directory") {
			t.Errorf("error = %v", err)
		}
	})
}

func TestGetExecutableDir(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := GetExecutableDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Dir(executable) {
		t.Errorf("directory = %q, executable = %q", dir, executable)
	}
}

func TestFindConfigFilePrecedence(t *testing.T) {
	isolateHome(t)
	exeDir, err := GetExecutableDir()
	if err != nil {
		t.Fatal(err)
	}
	// A unique file beside Go's temporary test executable exercises the real fallback.
	f, err := os.CreateTemp(exeDir, "xdg-discovery-*.json")
	if err != nil {
		t.Fatal(err)
	}
	exePath := f.Name()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(exePath); err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
	})
	filename := filepath.Base(exePath)
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	xdgPath := filepath.Join(dir, filename)
	if err := os.WriteFile(xdgPath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	explicit := filepath.Join(t.TempDir(), "explicit.json")
	if err := os.WriteFile(explicit, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, specified, want string }{
		{"explicit wins", explicit, explicit},
		{"XDG wins over executable", "", xdgPath},
		{"missing explicit falls back", explicit + ".missing", xdgPath},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindConfigFile(filename, tc.specified)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("path = %q, want %q", got, tc.want)
			}
		})
	}
	if err := os.Remove(xdgPath); err != nil {
		t.Fatal(err)
	}
	got, err := FindConfigFile(filename, "")
	if err != nil || got != exePath {
		t.Fatalf("executable fallback = %q, error %v", got, err)
	}
	// An unavailable config directory must still allow executable-directory fallback.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("APPDATA", "")
	got, err = FindConfigFile(filename, "")
	if err != nil || got != exePath {
		t.Fatalf("fallback without home = %q, error %v", got, err)
	}
	if err := os.Remove(exePath); err != nil {
		t.Fatal(err)
	}
	if _, err := FindConfigFile(filename, ""); err == nil || !strings.Contains(err.Error(), filename) {
		t.Errorf("missing file error = %v", err)
	}
}
