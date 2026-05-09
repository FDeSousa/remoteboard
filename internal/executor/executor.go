// Package executor provides an interface for executing button actions.
// Platform-specific implementations are in the sibling files (macos.go,
// linux.go, windows.go) and are selected at compile time via build tags.
package executor

import (
	"fmt"

	"github.com/FDeSousa/remoteboard/internal/config"
)

// Executor runs the action associated with a button.
type Executor interface {
	// Execute carries out the given action. It returns an error when the
	// action could not be performed (e.g. key not recognised, command
	// failed, etc.).
	Execute(action config.Action) error
}

// New returns the platform executor for the current OS (selected via build
// tags in the sibling files).
func New() Executor {
	return newPlatformExecutor()
}

// RunAction is a convenience wrapper that calls Execute on a new Executor.
func RunAction(action config.Action) error {
	return New().Execute(action)
}

// ErrUnknownActionType is returned when the action type is not recognised.
type ErrUnknownActionType struct {
	Type string
}

func (e *ErrUnknownActionType) Error() string {
	return fmt.Sprintf("executor: unknown action type %q", e.Type)
}
