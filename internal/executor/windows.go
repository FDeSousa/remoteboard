//go:build windows

package executor

import (
	"fmt"

	"github.com/FDeSousa/remoteboard/internal/config"
)

type windowsExecutor struct{}

func newPlatformExecutor() Executor { return &windowsExecutor{} }

func (e *windowsExecutor) Execute(action config.Action) error {
	switch action.Type {
	case "keypress":
		return fmt.Errorf("executor: keypress not yet implemented on Windows")
	case "shell":
		return e.shell(action.Command)
	default:
		return &ErrUnknownActionType{Type: action.Type}
	}
}

func (e *windowsExecutor) shell(command string) error {
	if command == "" {
		return fmt.Errorf("executor: shell: command must not be empty")
	}
	// TODO: implement via os/exec + cmd.exe
	return fmt.Errorf("executor: shell not yet implemented on Windows")
}
