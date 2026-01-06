package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Display    DisplayConfig    `toml:"display"`
	Appearance AppearanceConfig `toml:"appearance"`
	Behavior   BehaviorConfig   `toml:"behavior"`
	Privacy    PrivacyConfig    `toml:"privacy"`
}

type DisplayConfig struct {
	Position         string `toml:"position"` // top-left, top-right, bottom-left, bottom-right
	MarginX          int    `toml:"margin_x"`
	MarginY          int    `toml:"margin_y"`
	TimeoutMs        int    `toml:"timeout_ms"`
	HeldKeyTimeoutMs int    `toml:"held_key_timeout_ms"`
	HistoryCount     int    `toml:"history_count"`
	ShowHeldKeys     bool   `toml:"show_held_keys"`
}

type AppearanceConfig struct {
	Theme        string  `toml:"theme"` // dark, light
	FontSize     int     `toml:"font_size"`
	Opacity      float64 `toml:"opacity"`
	CornerRadius int     `toml:"corner_radius"`
}

type BehaviorConfig struct {
	CombineModifiers bool     `toml:"combine_modifiers"`
	ShowModifierOnly bool     `toml:"show_modifier_only"`
	ExcludedKeys     []string `toml:"excluded_keys"`
}

type PrivacyConfig struct {
	PauseOnApps []string `toml:"pause_on_apps"`
}

func Default() *Config {
	return &Config{
		Display: DisplayConfig{
			Position:         "bottom-right",
			MarginX:          20,
			MarginY:          40,
			TimeoutMs:        2000,
			HeldKeyTimeoutMs: 500,
			HistoryCount:     4,
			ShowHeldKeys:     true,
		},
		Appearance: AppearanceConfig{
			Theme:        "dark",
			FontSize:     18,
			Opacity:      0.85,
			CornerRadius: 8,
		},
		Behavior: BehaviorConfig{
			CombineModifiers: true,
			ShowModifierOnly: false,
			ExcludedKeys:     []string{},
		},
		Privacy: PrivacyConfig{
			PauseOnApps: []string{},
		},
	}
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return Default(), nil
	}

	return LoadFrom(path)
}

func LoadFrom(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}

	return c.SaveTo(path)
}

func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	return nil
}

func Path() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir: %w", err)
	}

	return filepath.Join(configDir, "tapshow", "config.toml"), nil
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.Display.TimeoutMs) * time.Millisecond
}

func (c *Config) HeldKeyTimeout() time.Duration {
	return time.Duration(c.Display.HeldKeyTimeoutMs) * time.Millisecond
}
