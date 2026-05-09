// remoteboard is the host agent for RemoteBoard.
//
// It:
//   - loads (or generates) the button config from ~/.config/remoteboard/buttons.json
//   - starts an HTTP + WebSocket server on the LAN
//   - watches the config file for live reloads
//   - on macOS: shows a system-tray / menu-bar icon with quick actions
//   - on other platforms: runs headlessly, printing the tablet URL to stdout
//
// Usage:
//
//remoteboard [flags]
//remoteboard install        # install launchd/systemd startup service
//remoteboard uninstall      # remove startup service
//remoteboard version        # print version and exit
//
// Flags:
//
//-addr   <host:port>   bind address (default: 0.0.0.0:8765)
//-config <path>        path to buttons.json (default: ~/.config/remoteboard/buttons.json)
package main

import (
"flag"
"fmt"
"log"
"os"

"github.com/FDeSousa/remoteboard/internal/config"
"github.com/FDeSousa/remoteboard/internal/server"
)

func main() {
if len(os.Args) > 1 {
switch os.Args[1] {
case "install":
if err := installService(); err != nil {
log.Fatalf("install: %v", err)
}
fmt.Println("RemoteBoard service installed. It will start at login.")
return
case "uninstall":
if err := uninstallService(); err != nil {
log.Fatalf("uninstall: %v", err)
}
fmt.Println("RemoteBoard service removed.")
return
case "version":
fmt.Println("remoteboard", version)
return
}
}

var (
addr    = flag.String("addr", "0.0.0.0:8765", "bind address (host:port)")
cfgPath = flag.String("config", "", "path to buttons.json (default: ~/.config/remoteboard/buttons.json)")
)
flag.Parse()

if *cfgPath == "" {
p, err := config.DefaultPath()
if err != nil {
log.Fatalf("config: %v", err)
}
*cfgPath = p
}

if err := config.EnsureDefault(*cfgPath); err != nil {
log.Fatalf("config: %v", err)
}

cfg, err := config.Load(*cfgPath)
if err != nil {
log.Fatalf("config: %v", err)
}

srv := server.New(*addr, *cfgPath, cfg)

// Start the config watcher.
stopWatch, err := server.WatchConfig(*cfgPath, func(newCfg *config.Config) {
srv.UpdateConfig(newCfg)
})
if err != nil {
log.Printf("watcher: could not watch config file: %v (live reload disabled)", err)
} else {
defer stopWatch()
}

// Determine the LAN URL for display / clipboard.
_, port, _ := parseHostPort(*addr)
tabletURL := "http://" + server.LANAddress(port)

// Start HTTP server in background; hand off to platform-specific run loop.
go func() {
if err := srv.ListenAndServe(); err != nil {
log.Fatalf("server: %v", err)
}
}()

runApp(tabletURL, *cfgPath)
}

// parseHostPort splits "host:port" and returns host and port strings.
func parseHostPort(addr string) (host, port string, err error) {
for i := len(addr) - 1; i >= 0; i-- {
if addr[i] == ':' {
return addr[:i], addr[i+1:], nil
}
}
return "", addr, fmt.Errorf("no port in address %q", addr)
}
