//go:build darwin

package executor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/FDeSousa/remoteboard/internal/config"
)

type macosExecutor struct{}

func newPlatformExecutor() Executor { return &macosExecutor{} }

func (e *macosExecutor) Execute(action config.Action) error {
	switch action.Type {
	case "keypress":
		return e.keypress(action.Key, action.Modifiers)
	case "shell":
		return e.shell(action.Command)
	default:
		return &ErrUnknownActionType{Type: action.Type}
	}
}

// keypress sends a keystroke via osascript so that no CGo or extra
// permissions beyond Accessibility are required.
func (e *macosExecutor) keypress(key string, modifiers []string) error {
	if key == "" {
		return fmt.Errorf("executor: keypress: key must not be empty")
	}

	// Separate function keys (F1–F20) from regular keys because osascript
	// uses "key code" for those and "keystroke" for character keys.
	if code, ok := fKeyCode(key); ok {
		return e.keyCode(code, modifiers)
	}
	return e.keystroke(key, modifiers)
}

// keystroke sends a regular character keystroke.
func (e *macosExecutor) keystroke(key string, modifiers []string) error {
	using := buildUsing(modifiers)
	var script string
	if using == "" {
		script = fmt.Sprintf(
			`tell application "System Events" to keystroke %q`,
			key,
		)
	} else {
		script = fmt.Sprintf(
			`tell application "System Events" to keystroke %q using {%s}`,
			key, using,
		)
	}
	return runOsascript(script)
}

// keyCode sends a key-code-based keystroke (used for function keys, etc.).
func (e *macosExecutor) keyCode(code int, modifiers []string) error {
	using := buildUsing(modifiers)
	var script string
	if using == "" {
		script = fmt.Sprintf(
			`tell application "System Events" to key code %d`,
			code,
		)
	} else {
		script = fmt.Sprintf(
			`tell application "System Events" to key code %d using {%s}`,
			code, using,
		)
	}
	return runOsascript(script)
}

// shell executes an arbitrary shell command via /bin/sh.
func (e *macosExecutor) shell(command string) error {
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

// runOsascript runs an AppleScript expression via the osascript binary.
func runOsascript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("executor: osascript: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// buildUsing converts modifier names to the AppleScript "using" list syntax.
// Supported names: cmd / command, shift, ctrl / control, alt / option.
func buildUsing(modifiers []string) string {
	var parts []string
	for _, m := range modifiers {
		switch strings.ToLower(m) {
		case "cmd", "command":
			parts = append(parts, "command down")
		case "shift":
			parts = append(parts, "shift down")
		case "ctrl", "control":
			parts = append(parts, "control down")
		case "alt", "option":
			parts = append(parts, "option down")
		}
	}
	return strings.Join(parts, ", ")
}

// fKeyCode maps function-key names to their macOS key codes.
func fKeyCode(key string) (int, bool) {
	codes := map[string]int{
		"F1": 122, "F2": 120, "F3": 99, "F4": 118,
		"F5": 96, "F6": 97, "F7": 98, "F8": 100,
		"F9": 101, "F10": 109, "F11": 103, "F12": 111,
		"F13": 105, "F14": 107, "F15": 113, "F16": 106,
		"F17": 64, "F18": 79, "F19": 80, "F20": 90,
	}
	code, ok := codes[strings.ToUpper(key)]
	return code, ok
}
