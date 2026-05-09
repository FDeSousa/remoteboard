// Package config handles loading, saving, and watching the RemoteBoard
// configuration file (~/.config/remoteboard/buttons.json).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Version is the only supported config file version.
const Version = 1

// Grid holds layout options for the button grid.
type Grid struct {
	Columns int `json:"columns"`
}

// Action describes what happens when a button is pressed.
type Action struct {
	// Type is either "keypress" or "shell".
	Type string `json:"type"`

	// Keypress fields.
	Key       string   `json:"key,omitempty"`
	Modifiers []string `json:"modifiers,omitempty"`

	// Shell fields.
	Command string `json:"command,omitempty"`
}

// Button is a single tile displayed in the grid.
type Button struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Icon   string `json:"icon,omitempty"`  // SF Symbol name
	Color  string `json:"color,omitempty"` // hex colour, e.g. "#e74c3c"
	Action Action `json:"action"`
}

// Page groups a set of buttons. The data model supports multiple pages from
// day one; the MVP UI only renders the first page.
type Page struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Buttons []Button `json:"buttons"`
}

// Config is the top-level structure of buttons.json.
type Config struct {
	Version int    `json:"version"`
	Grid    Grid   `json:"grid"`
	Pages   []Page `json:"pages"`
	// PIN is optional. When non-empty, POST /api/trigger/:id requires the
	// header "X-RemoteBoard-PIN" to match this value.
	PIN string `json:"pin,omitempty"`
}

// DefaultPath returns the platform-conventional path for buttons.json.
func DefaultPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config: cannot determine user config dir: %w", err)
	}
	return filepath.Join(cfgDir, "remoteboard", "buttons.json"), nil
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if cfg.Version != Version {
		return nil, fmt.Errorf("config: unsupported version %d (expected %d)", cfg.Version, Version)
	}
	return &cfg, nil
}

// Save writes cfg as formatted JSON to path, creating parent directories as
// needed.
func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("config: mkdir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("config: write %s: %w", path, err)
	}
	return nil
}

// EnsureDefault creates a default config file at path if it does not already
// exist.
func EnsureDefault(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config: stat %s: %w", path, err)
	}
	return Save(path, defaultConfig())
}

// ButtonByID returns the button with the given id across all pages, or an
// error if not found.
func (c *Config) ButtonByID(id string) (Button, error) {
	for _, page := range c.Pages {
		for _, btn := range page.Buttons {
			if btn.ID == id {
				return btn, nil
			}
		}
	}
	return Button{}, fmt.Errorf("config: button %q not found", id)
}

// defaultConfig returns a sensible starter config with a handful of useful
// macOS shortcuts.
func defaultConfig() *Config {
	return &Config{
		Version: Version,
		Grid:    Grid{Columns: 4},
		Pages: []Page{
			{
				ID:    "default",
				Label: "Main",
				Buttons: []Button{
					{
						ID:    "mute",
						Label: "Mute",
						Icon:  "mic.slash",
						Color: "#e74c3c",
						Action: Action{
							Type:      "keypress",
							Key:       "F10",
							Modifiers: []string{},
						},
					},
					{
						ID:    "screenshot",
						Label: "Screenshot",
						Icon:  "camera",
						Color: "#3498db",
						Action: Action{
							Type:      "keypress",
							Key:       "4",
							Modifiers: []string{"cmd", "shift"},
						},
					},
					{
						ID:    "terminal",
						Label: "Terminal",
						Icon:  "terminal",
						Color: "#2ecc71",
						Action: Action{
							Type:    "shell",
							Command: "open -a Terminal",
						},
					},
					{
						ID:    "lock",
						Label: "Lock",
						Icon:  "lock",
						Color: "#95a5a6",
						Action: Action{
							Type:      "keypress",
							Key:       "q",
							Modifiers: []string{"cmd", "ctrl"},
						},
					},
					{
						ID:    "spotlight",
						Label: "Spotlight",
						Icon:  "magnifyingglass",
						Color: "#9b59b6",
						Action: Action{
							Type:      "keypress",
							Key:       "space",
							Modifiers: []string{"cmd"},
						},
					},
					{
						ID:    "dnd",
						Label: "Do Not Disturb",
						Icon:  "moon",
						Color: "#f39c12",
						Action: Action{
							Type:    "shell",
							Command: `osascript -e 'tell application "System Events" to key code 64 using {option down}'`,
						},
					},
				},
			},
		},
	}
}
