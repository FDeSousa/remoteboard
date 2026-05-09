package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FDeSousa/remoteboard/internal/config"
	"github.com/FDeSousa/remoteboard/internal/server"
)

// newTestServer returns a Server wired to a test config, with a test
// HTTP server that is closed automatically when the test finishes.
func newTestServer(t *testing.T, cfg *config.Config) *httptest.Server {
	t.Helper()
	srv := server.New("127.0.0.1:0", "", cfg)
	return httptest.NewServer(srv.Handler())
}

func testConfig() *config.Config {
	return &config.Config{
		Version: config.Version,
		Grid:    config.Grid{Columns: 4},
		Pages: []config.Page{
			{
				ID:    "default",
				Label: "Main",
				Buttons: []config.Button{
					{
						ID:    "noop",
						Label: "No-op",
						Action: config.Action{
							Type:    "shell",
							Command: "true", // always succeeds
						},
					},
				},
			},
		},
	}
}

func TestHandleHealth(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestHandleButtons(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/buttons")
	if err != nil {
		t.Fatalf("GET /api/buttons: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var cfg config.Config
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(cfg.Pages) == 0 {
		t.Error("response has no pages")
	}
}

func TestHandleTriggerNotFound(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/trigger/nonexistent", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatalf("POST /api/trigger/nonexistent: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleTriggerMissingID(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/trigger/", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatalf("POST /api/trigger/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleTriggerMethodNotAllowed(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/trigger/noop")
	if err != nil {
		t.Fatalf("GET /api/trigger/noop: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestPINAuth(t *testing.T) {
	cfg := testConfig()
	cfg.PIN = "secret"
	ts := newTestServer(t, cfg)
	defer ts.Close()

	// Without PIN — should be rejected.
	resp, err := http.Post(ts.URL+"/api/trigger/noop", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("without PIN: status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	// With wrong PIN — should be rejected.
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/trigger/noop", strings.NewReader(""))
	req.Header.Set("X-RemoteBoard-PIN", "wrong")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong PIN: status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestCORSHeaders(t *testing.T) {
	ts := newTestServer(t, testConfig())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}

func TestLANAddress(t *testing.T) {
	addr := server.LANAddress("8765")
	if addr == "" {
		t.Error("LANAddress returned empty string")
	}
	if !strings.Contains(addr, ":8765") {
		t.Errorf("LANAddress = %q, want it to contain :8765", addr)
	}
}
