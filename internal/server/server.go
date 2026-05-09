// Package server provides the HTTP and WebSocket server for RemoteBoard.
//
// Endpoints:
//
//GET  /api/buttons       – full config (all pages)
//POST /api/trigger/:id   – execute button action (optional PIN auth)
//GET  /api/health        – heartbeat / liveness
//GET  /ws                – WebSocket; receives config-reload events
package server

import (
"encoding/json"
"fmt"
"log"
"net"
"net/http"
"strings"
"sync"
"time"

"github.com/FDeSousa/remoteboard/internal/config"
"github.com/FDeSousa/remoteboard/internal/executor"
"github.com/gorilla/websocket"
)

// Server is the RemoteBoard HTTP server.
type Server struct {
addr     string
cfgPath  string
cfg      *config.Config
cfgMu    sync.RWMutex
exec     executor.Executor
upgrader websocket.Upgrader
clients  map[*websocket.Conn]struct{}
clientMu sync.Mutex
handler  http.Handler
}

// New creates a Server bound to addr, serving configuration from cfgPath.
func New(addr, cfgPath string, cfg *config.Config) *Server {
s := &Server{
addr:    addr,
cfgPath: cfgPath,
cfg:     cfg,
exec:    executor.New(),
clients: make(map[*websocket.Conn]struct{}),
upgrader: websocket.Upgrader{
CheckOrigin: func(r *http.Request) bool { return true },
},
}
s.handler = s.buildHandler()
return s
}

// buildHandler wires up the HTTP routes and wraps them with CORS middleware.
func (s *Server) buildHandler() http.Handler {
mux := http.NewServeMux()
mux.HandleFunc("/api/buttons", s.handleButtons)
mux.HandleFunc("/api/trigger/", s.handleTrigger)
mux.HandleFunc("/api/health", s.handleHealth)
mux.HandleFunc("/ws", s.handleWS)
return corsMiddleware(mux)
}

// Handler returns the http.Handler for use in tests or custom servers.
func (s *Server) Handler() http.Handler { return s.handler }

// UpdateConfig replaces the in-memory config and notifies all WebSocket clients.
func (s *Server) UpdateConfig(cfg *config.Config) {
s.cfgMu.Lock()
s.cfg = cfg
s.cfgMu.Unlock()
s.broadcast(`{"event":"config_reload"}`)
}

// ListenAndServe starts the HTTP server and blocks until it returns an error.
func (s *Server) ListenAndServe() error {
srv := &http.Server{
Addr:         s.addr,
Handler:      s.handler,
ReadTimeout:  15 * time.Second,
WriteTimeout: 15 * time.Second,
IdleTimeout:  60 * time.Second,
}
log.Printf("server: listening on %s", s.addr)
return srv.ListenAndServe()
}

// --- handlers ---

func (s *Server) handleButtons(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
s.cfgMu.RLock()
cfg := s.cfg
s.cfgMu.RUnlock()

writeJSON(w, cfg)
}

func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

// Authenticate via optional PIN.
s.cfgMu.RLock()
cfg := s.cfg
s.cfgMu.RUnlock()

if cfg.PIN != "" {
pin := r.Header.Get("X-RemoteBoard-PIN")
if pin != cfg.PIN {
http.Error(w, "unauthorized", http.StatusUnauthorized)
return
}
}

// Extract button ID from path: /api/trigger/<id>
id := strings.TrimPrefix(r.URL.Path, "/api/trigger/")
if id == "" {
http.Error(w, "button id required", http.StatusBadRequest)
return
}

btn, err := cfg.ButtonByID(id)
if err != nil {
http.Error(w, fmt.Sprintf("button %q not found", id), http.StatusNotFound)
return
}

if err := s.exec.Execute(btn.Action); err != nil {
log.Printf("server: trigger %q: %v", id, err)
http.Error(w, "action failed: "+err.Error(), http.StatusInternalServerError)
return
}

w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
conn, err := s.upgrader.Upgrade(w, r, nil)
if err != nil {
log.Printf("server: ws upgrade: %v", err)
return
}

s.clientMu.Lock()
s.clients[conn] = struct{}{}
s.clientMu.Unlock()

log.Printf("server: ws client connected: %s", conn.RemoteAddr())

// Read loop — we only need to handle pings and detect disconnections.
go func() {
defer func() {
s.clientMu.Lock()
delete(s.clients, conn)
s.clientMu.Unlock()
conn.Close()
log.Printf("server: ws client disconnected: %s", conn.RemoteAddr())
}()
for {
if _, _, err := conn.ReadMessage(); err != nil {
return
}
}
}()
}

// broadcast sends a text message to all connected WebSocket clients.
func (s *Server) broadcast(msg string) {
s.clientMu.Lock()
defer s.clientMu.Unlock()
for conn := range s.clients {
if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
log.Printf("server: ws broadcast: %v", err)
conn.Close()
delete(s.clients, conn)
}
}
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, v any) {
w.Header().Set("Content-Type", "application/json")
enc := json.NewEncoder(w)
enc.SetIndent("", "  ")
if err := enc.Encode(v); err != nil {
log.Printf("server: json encode: %v", err)
}
}

func corsMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-RemoteBoard-PIN")
if r.Method == http.MethodOptions {
w.WriteHeader(http.StatusNoContent)
return
}
next.ServeHTTP(w, r)
})
}

// LANAddress returns the first non-loopback IPv4 address suitable for
// displaying to the user ("connect your tablet to http://<ip>:<port>").
func LANAddress(port string) string {
ifaces, err := net.Interfaces()
if err != nil {
return "localhost:" + port
}
for _, iface := range ifaces {
if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
continue
}
addrs, err := iface.Addrs()
if err != nil {
continue
}
for _, addr := range addrs {
var ip net.IP
switch v := addr.(type) {
case *net.IPNet:
ip = v.IP
case *net.IPAddr:
ip = v.IP
}
if ip == nil || ip.IsLoopback() {
continue
}
if ip4 := ip.To4(); ip4 != nil {
return ip4.String() + ":" + port
}
}
}
return "localhost:" + port
}
