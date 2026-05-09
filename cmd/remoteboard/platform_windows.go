//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

const version = "0.1.0"

func copyToClipboard(text string) {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}

func openFile(path string) {
	_ = exec.Command("explorer", path).Start()
}

func installService() error {
	return fmt.Errorf("install: not yet implemented on Windows")
}

func uninstallService() error {
	return fmt.Errorf("uninstall: not yet implemented on Windows")
}
