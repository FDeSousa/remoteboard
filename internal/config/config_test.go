package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FDeSousa/remoteboard/internal/config"
)

func TestDefaultPath(t *testing.T) {
	p, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}
	if p == "" {
		t.Fatal("DefaultPath() returned empty string")
	}
}

func TestEnsureDefaultAndLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "buttons.json")

	// Should create the file.
	if err := config.EnsureDefault(p); err != nil {
		t.Fatalf("EnsureDefault() error: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	// Should load a valid config.
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Version != config.Version {
		t.Errorf("Version = %d, want %d", cfg.Version, config.Version)
	}
	if len(cfg.Pages) == 0 {
		t.Error("no pages in default config")
	}
	if len(cfg.Pages[0].Buttons) == 0 {
		t.Error("no buttons in default page")
	}

	// EnsureDefault should be idempotent (does not overwrite).
	if err := config.EnsureDefault(p); err != nil {
		t.Fatalf("EnsureDefault() second call error: %v", err)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "buttons.json")

	original := &config.Config{
		Version: config.Version,
		Grid:    config.Grid{Columns: 3},
		Pages: []config.Page{
			{
				ID:    "test",
				Label: "Test",
				Buttons: []config.Button{
					{
						ID:    "btn1",
						Label: "Button 1",
						Action: config.Action{
							Type: "shell",
							Command: "echo hello",
						},
					},
				},
			},
		},
	}

	if err := config.Save(p, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := config.Load(p)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Grid.Columns != 3 {
		t.Errorf("Grid.Columns = %d, want 3", loaded.Grid.Columns)
	}
	if len(loaded.Pages) != 1 {
		t.Fatalf("len(Pages) = %d, want 1", len(loaded.Pages))
	}
	if loaded.Pages[0].Buttons[0].ID != "btn1" {
		t.Errorf("Button ID = %q, want %q", loaded.Pages[0].Buttons[0].ID, "btn1")
	}
}

func TestButtonByID(t *testing.T) {
	cfg := &config.Config{
		Version: config.Version,
		Pages: []config.Page{
			{
				ID: "p1",
				Buttons: []config.Button{
					{ID: "a", Label: "A"},
					{ID: "b", Label: "B"},
				},
			},
		},
	}

	btn, err := cfg.ButtonByID("b")
	if err != nil {
		t.Fatalf("ButtonByID() error: %v", err)
	}
	if btn.Label != "B" {
		t.Errorf("Label = %q, want %q", btn.Label, "B")
	}

	_, err = cfg.ButtonByID("missing")
	if err == nil {
		t.Error("ButtonByID(missing) should return error")
	}
}

func TestLoadInvalidVersion(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte(`{"version":99,"grid":{"columns":4},"pages":[]}`), 0o640); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(p)
	if err == nil {
		t.Error("Load() with wrong version should return error")
	}
}
