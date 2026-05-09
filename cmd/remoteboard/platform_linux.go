//go:build linux

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

// copyToClipboard copies text to the clipboard via xclip or xsel.
func copyToClipboard(text string) {
	for _, tool := range []string{"xclip", "xsel"} {
		cmd := exec.Command(tool)
		if tool == "xclip" {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return
		}
	}
}

// openFile opens the file at path using xdg-open.
func openFile(path string) {
	_ = exec.Command("xdg-open", path).Start()
}

// --- systemd user service installer ---

const systemdUnitTemplate = `[Unit]
Description=RemoteBoard Host Agent
After=network.target

[Service]
ExecStart={{.BinaryPath}}
Restart=on-failure

[Install]
WantedBy=default.target
`

func unitPath() string {
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		home, _ := os.UserHomeDir()
		cfgHome = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgHome, "systemd", "user", "remoteboard.service")
}

func installService() error {
	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	dst := unitPath()
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("unit").Parse(systemdUnitTemplate))
	data := struct{ BinaryPath string }{BinaryPath: bin}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	if out, err := exec.Command("systemctl", "--user", "enable", "--now", "remoteboard.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable: %w (output: %s)", err, string(out))
	}
	return nil
}

func uninstallService() error {
	dst := unitPath()
	if out, err := exec.Command("systemctl", "--user", "disable", "--now", "remoteboard.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl disable: %w (output: %s)", err, string(out))
	}
	return os.Remove(dst)
}
