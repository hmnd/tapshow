package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Display.Position != "bottom-right" {
		t.Errorf("Default Position = %q, want 'bottom-right'", cfg.Display.Position)
	}
	if cfg.Display.HistoryCount != 4 {
		t.Errorf("Default HistoryCount = %d, want 4", cfg.Display.HistoryCount)
	}

	if cfg.Appearance.Theme != "dark" {
		t.Errorf("Default Theme = %q, want 'dark'", cfg.Appearance.Theme)
	}
	if cfg.Appearance.FontSize != 18 {
		t.Errorf("Default FontSize = %d, want 18", cfg.Appearance.FontSize)
	}
	if cfg.Appearance.Opacity != 0.85 {
		t.Errorf("Default Opacity = %f, want 0.85", cfg.Appearance.Opacity)
	}

	if !cfg.Behavior.CombineModifiers {
		t.Error("Default CombineModifiers should be true")
	}
}

func TestTimeout(t *testing.T) {
	cfg := Default()
	cfg.Display.TimeoutMs = 2000

	timeout := cfg.Timeout()
	expected := 2000 * time.Millisecond

	if timeout != expected {
		t.Errorf("Timeout() = %v, want %v", timeout, expected)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := Default()
	cfg.Display.HistoryCount = 10
	cfg.Appearance.Theme = "light"
	cfg.Appearance.FontSize = 24
	cfg.Behavior.CombineModifiers = false

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Display.HistoryCount != 10 {
		t.Errorf("Loaded HistoryCount = %d, want 10", loaded.Display.HistoryCount)
	}
	if loaded.Appearance.Theme != "light" {
		t.Errorf("Loaded Theme = %q, want 'light'", loaded.Appearance.Theme)
	}
	if loaded.Appearance.FontSize != 24 {
		t.Errorf("Loaded FontSize = %d, want 24", loaded.Appearance.FontSize)
	}
	if loaded.Behavior.CombineModifiers {
		t.Error("Loaded CombineModifiers should be false")
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("LoadFrom should not error for non-existent file: %v", err)
	}

	if cfg.Display.HistoryCount != 4 {
		t.Errorf("Should have default HistoryCount, got %d", cfg.Display.HistoryCount)
	}
}

func TestPath(t *testing.T) {
	path, err := Path()
	if err != nil {
		t.Fatalf("Path() failed: %v", err)
	}

	if filepath.Base(path) != "config.toml" {
		t.Errorf("Path should end with config.toml, got %q", path)
	}

	dir := filepath.Dir(path)
	if filepath.Base(dir) != "tapshow" {
		t.Errorf("Config should be in tapshow dir, got %q", dir)
	}
}

func TestPrivacyConfig(t *testing.T) {
	cfg := Default()

	if len(cfg.Privacy.PauseOnApps) != 0 {
		t.Errorf("Default PauseOnApps should be empty, got %v", cfg.Privacy.PauseOnApps)
	}

	cfg.Privacy.PauseOnApps = []string{"1password", "keepassxc"}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if len(loaded.Privacy.PauseOnApps) != 2 {
		t.Errorf("Loaded PauseOnApps length = %d, want 2", len(loaded.Privacy.PauseOnApps))
	}
}

func TestExcludedKeys(t *testing.T) {
	cfg := Default()
	cfg.Behavior.ExcludedKeys = []string{"CapsLock", "NumLock"}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if len(loaded.Behavior.ExcludedKeys) != 2 {
		t.Errorf("Loaded ExcludedKeys length = %d, want 2", len(loaded.Behavior.ExcludedKeys))
	}
}
