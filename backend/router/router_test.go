package router

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	cfg := &config.Config{
		Port:         "0",
		DatabasePath: dbPath,
		Auth: config.AuthConfig{
			Username: "admin",
			Password: "changeme",
		},
		RateLimit: config.RateLimitConfig{
			Enabled: true,
			Storage: "memory",
			DefaultRule: config.RateLimitRule{
				Enabled:        true,
				BucketSize:     1000,
				RefillSize:     0,
				RefillDuration: time.Hour,
				IdentifyBy:     "ip",
			},
			Routes: map[string]config.RateLimitRule{},
		},
	}

	if err := storage.InitDB(cfg); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		_ = storage.Close()
	})

	return cfg
}

func authHeader(user, pass string) string {
	b := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + b
}

func TestHealth_OK(t *testing.T) {
	cfg := newTestConfig(t)
	h := Init(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body)
	}
}

func TestRuns_LegacyUnchanged(t *testing.T) {
	cfg := newTestConfig(t)
	if err := storage.SaveRun(&storage.Run{Iteration: 1, Score: 10, Weakness: "specificity"}); err != nil {
		t.Fatalf("SaveRun failed: %v", err)
	}

	h := Init(cfg)
	req := httptest.NewRequest(http.MethodGet, "/runs", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Runs []map[string]any `json:"runs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(out.Runs))
	}
}

func TestRun_Unauthorized(t *testing.T) {
	cfg := newTestConfig(t)
	h := Init(cfg)

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewBufferString(`{"code":"package main\nfunc main(){}"}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRun_BadRequest_InvalidJSON(t *testing.T) {
	cfg := newTestConfig(t)
	h := Init(cfg)

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewBufferString(`{"code":`))
	req.Header.Set("Authorization", authHeader(cfg.Auth.Username, cfg.Auth.Password))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRun_AsyncStarted(t *testing.T) {
	cfg := newTestConfig(t)
	h := Init(cfg)

	// Stub the async processor so tests are fast and deterministic.
	old := processRunGroupAsync
	processRunGroupAsync = func(cfg *config.Config, runGroupID int, code string, extraPrompt string) {
		_ = storage.UpdateRunGroupStatus(runGroupID, "completed")
		_ = storage.SaveRunGroupRun(runGroupID, 1, 9, "actionability")
	}
	t.Cleanup(func() { processRunGroupAsync = old })

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewBufferString(`{"code":"package main\n\nfunc main() {}\n"}`))
	req.Header.Set("Authorization", authHeader(cfg.Auth.Username, cfg.Auth.Password))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["status"] != "started" {
		t.Fatalf("expected status started, got %v", out["status"])
	}
	if _, ok := out["run_group_id"]; !ok {
		t.Fatalf("expected run_group_id in response")
	}
}

func TestRunGroups_Unauthorized(t *testing.T) {
	cfg := newTestConfig(t)
	h := Init(cfg)

	req := httptest.NewRequest(http.MethodGet, "/run-groups?limit=10&offset=0", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRateLimit_TooManyRequests(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.RateLimit.DefaultRule = config.RateLimitRule{
		Enabled:        true,
		BucketSize:     1,
		RefillSize:     0,
		RefillDuration: time.Hour,
		IdentifyBy:     "ip",
	}

	h := Init(cfg)
	req1 := httptest.NewRequest(http.MethodGet, "/runs", nil)
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d: %s", rr1.Code, rr1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/runs", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d: %s", rr2.Code, rr2.Body.String())
	}
}
