//go:build linux

package executor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/FDeSousa/remoteboard/internal/config"
)

type linuxExecutor struct{}

func newPlatformExecutor() Executor { return &linuxExecutor{} }

func (e *linuxExecutor) Execute(action config.Action) error {
	switch action.Type {
	case "keypress":
		return e.keypress(action.Key, action.Modifiers)
	case "shell":
		return e.shell(action.Command)
	default:
		return &ErrUnknownActionType{Type: action.Type}
	}
}

// keypress sends a keystroke via xdotool.
func (e *linuxExecutor) keypress(key string, modifiers []string) error {
	if key == "" {
		return fmt.Errorf("executor: keypress: key must not be empty")
	}

	combo := buildXdotoolCombo(key, modifiers)
	cmd := exec.Command("xdotool", "key", combo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("executor: xdotool: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// shell executes an arbitrary shell command via /bin/sh.
func (e *linuxExecutor) shell(command string) error {
	if command == "" {
		return fmt.Errorf("executor: shell: command must not be empty")
	}
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("executor: shell: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// buildXdotoolCombo builds an xdotool key combo string (e.g. "ctrl+shift+s").
func buildXdotoolCombo(key string, modifiers []string) string {
	var parts []string
	for _, m := range modifiers {
		switch strings.ToLower(m) {
		case "cmd", "command":
			parts = append(parts, "super")
		case "shift":
			parts = append(parts, "shift")
		case "ctrl", "control":
			parts = append(parts, "ctrl")
		case "alt", "option":
			parts = append(parts, "alt")
		}
	}
	parts = append(parts, key)
	return strings.Join(parts, "+")
}
