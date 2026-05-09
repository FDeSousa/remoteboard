//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const version = "0.1.0"

// copyToClipboard copies text to the macOS clipboard via pbcopy.
func copyToClipboard(text string) {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		// Non-fatal; the URL is visible in the menu item label anyway.
		_ = err
	}
}

// openFile opens the file at path in the default macOS application.
func openFile(path string) {
	_ = exec.Command("open", path).Start()
}

// --- launchd service installer ---

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.remoteboard.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/remoteboard.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/remoteboard.log</string>
</dict>
</plist>
`

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.remoteboard.agent.plist")
}

func logDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Logs")
}

func installService() error {
	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	dst := plistPath()
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("plist").Parse(launchdPlistTemplate))
	data := struct {
		BinaryPath string
		LogDir     string
	}{BinaryPath: bin, LogDir: logDir()}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	// Load the plist immediately.
	out, err := exec.Command("launchctl", "load", dst).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %w (output: %s)", err, string(out))
	}
	return nil
}

func uninstallService() error {
	dst := plistPath()
	out, err := exec.Command("launchctl", "unload", dst).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl unload: %w (output: %s)", err, string(out))
	}
	return os.Remove(dst)
}
