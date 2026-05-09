# RemoteBoard

A lightweight remote button board for tablets and phones. Run the Go host agent on your Mac (or Linux/Windows), then connect from any browser or the native Swift/Skip.dev app to trigger keyboard shortcuts and shell commands with a tap.

```
┌─────────────────────────┐        LAN (HTTP/WebSocket)       ┌────────────────────────┐
│  Swift / Skip.dev App   │ ◄────────────────────────────────► │  Go Host Agent         │
│  (iOS + Android)        │                                    │  (macOS menu-bar app)  │
│                         │                                    │                        │
│  • Full-screen grid     │   GET  /api/buttons                │  • Serves tablet UI    │
│  • Wake lock            │   POST /api/trigger/:id            │  • Reads config file   │
│  • Connection status    │   WS   /ws  (config push)          │  • Simulates keypresses│
└─────────────────────────┘                                    └────────────────────────┘
```

---

## Requirements

- **Go 1.21+** (`go version`)
- **macOS** (primary target; Linux/Windows supported headlessly)
- macOS **Accessibility permission** for the `remoteboard` binary (System Settings → Privacy & Security → Accessibility) — required for keypress simulation via `osascript`

---

## Install

```sh
# From source
go install github.com/FDeSousa/remoteboard/cmd/remoteboard@latest

# Or clone and build
git clone https://github.com/FDeSousa/remoteboard.git
cd remoteboard
go build -o remoteboard ./cmd/remoteboard
```

---

## Quick start

```sh
# 1. Start the agent (creates a default config on first run)
remoteboard

# The menu-bar icon shows the tablet URL, e.g.:
#   Tablet URL: http://192.168.1.42:8765

# 2. Open that URL on your tablet in Safari / Chrome, or use the native app.

# 3. Tap a button — the action fires on your Mac instantly.
```

### Optional: start at login (macOS)

```sh
remoteboard install    # installs a launchd plist in ~/Library/LaunchAgents/
remoteboard uninstall  # removes it
```

---

## Configuration

The config file lives at `~/.config/remoteboard/buttons.json` (or the path passed with `-config`).

Edit it with any text editor; RemoteBoard watches for changes and reloads automatically — no restart needed.

### Example `buttons.json`

```json
{
  "version": 1,
  "grid": { "columns": 4 },
  "pages": [
    {
      "id": "default",
      "label": "Main",
      "buttons": [
        {
          "id": "mute",
          "label": "Mute",
          "icon": "mic.slash",
          "color": "#e74c3c",
          "action": {
            "type": "keypress",
            "key": "F10",
            "modifiers": []
          }
        },
        {
          "id": "screenshot",
          "label": "Screenshot",
          "icon": "camera",
          "color": "#3498db",
          "action": {
            "type": "keypress",
            "key": "4",
            "modifiers": ["cmd", "shift"]
          }
        },
        {
          "id": "terminal",
          "label": "Terminal",
          "icon": "terminal",
          "color": "#2ecc71",
          "action": {
            "type": "shell",
            "command": "open -a Terminal"
          }
        }
      ]
    }
  ]
}
```

### Button fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | ✓ | Unique identifier (used in POST /api/trigger/:id) |
| `label` | string | ✓ | Display name on the button |
| `icon` | string | | SF Symbol name (e.g. `"mic.slash"`, `"camera"`) |
| `color` | string | | Hex accent colour (e.g. `"#e74c3c"`) |
| `action` | object | ✓ | See below |

### Action types

#### `keypress`

Sends a keyboard shortcut via `osascript` (macOS) or `xdotool` (Linux).

```json
{
  "type": "keypress",
  "key": "4",
  "modifiers": ["cmd", "shift"]
}
```

Supported modifiers: `cmd` / `command`, `shift`, `ctrl` / `control`, `alt` / `option`.

Function keys (F1–F20) are supported as the `key` value.

#### `shell`

Runs an arbitrary shell command via `/bin/sh -c`.

```json
{
  "type": "shell",
  "command": "open -a Terminal"
}
```

### Optional PIN protection

Add a `"pin"` field to the root of `buttons.json`. When set, all `POST /api/trigger/:id` requests must include the header `X-RemoteBoard-PIN: <pin>`.

```json
{
  "version": 1,
  "pin": "1234",
  "grid": { "columns": 4 },
  "pages": [...]
}
```

---

## HTTP API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/buttons` | none | Full config (all pages) |
| `POST` | `/api/trigger/:id` | optional PIN | Execute button action |
| `GET` | `/api/health` | none | `{"status":"ok"}` heartbeat |
| `WS` | `/ws` | none | Push `{"event":"config_reload"}` when config changes |

---

## CLI flags

```
remoteboard [flags] [command]

Flags:
  -addr   host:port   Bind address (default: 0.0.0.0:8765)
  -config path        Path to buttons.json

Commands:
  install     Install launchd (macOS) / systemd (Linux) startup service
  uninstall   Remove startup service
  version     Print version
```

---

## Development

```sh
# Run tests
go test ./...

# Build for macOS (from macOS)
go build -o remoteboard ./cmd/remoteboard

# Build headless (Linux / CI)
go build -o remoteboard ./cmd/remoteboard
```

---

## Architecture

The project is organised as a standard Go module:

```
remoteboard/
  cmd/remoteboard/          # main entry point
    main.go                 # flag parsing, server start, config watcher
    app_darwin.go           # macOS menu-bar icon (systray)
    app_other.go            # headless run loop for Linux/Windows
    platform_darwin.go      # clipboard, launchd installer
    platform_linux.go       # clipboard, systemd installer
    platform_windows.go     # clipboard stub, Windows installer stub
  internal/
    config/                 # load/save/watch buttons.json
    executor/               # action executor interface + platform impls
    server/                 # HTTP + WebSocket server
```

Platform-specific code is gated by Go build tags (`darwin`, `linux`, `windows`) so the binary compiles cleanly on all targets.

---

## Roadmap

- [ ] Native Swift / Skip.dev tablet app (iOS + Android)
- [ ] Multi-page support in the tablet UI
- [ ] mDNS auto-discovery (Bonjour)
- [ ] HTTPS with self-signed cert
- [ ] Macro sequences (multiple actions per button)
- [ ] Config GUI (drag-and-drop editor in browser)
- [ ] Windows keypress executor (`SendInput`)
