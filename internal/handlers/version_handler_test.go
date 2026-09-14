package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The version endpoint must report what was linked into the binary, since the
// UI shows it in place of a hardcoded string that used to drift.
func TestGetVersionHandler(t *testing.T) {
	handler := &Handler{
		buildInfo: BuildInfo{
			Version: "v2.2.7-1-gcc75c75",
			Commit:  "cc75c75",
			Date:    "2026-09-10T08:03:49Z",
		},
	}

	recorder := httptest.NewRecorder()
	handler.getVersionHandler(recorder, httptest.NewRequest(http.MethodGet, "/api/version", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var got BuildInfo
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.Version != "v2.2.7-1-gcc75c75" {
		t.Errorf("version = %q, want %q", got.Version, "v2.2.7-1-gcc75c75")
	}
	if got.Commit != "cc75c75" {
		t.Errorf("commit = %q, want %q", got.Commit, "cc75c75")
	}
	if got.Date != "2026-09-10T08:03:49Z" {
		t.Errorf("date = %q, want %q", got.Date, "2026-09-10T08:03:49Z")
	}
}

// An unlinked build reports "dev" rather than an empty string, and the endpoint
// must pass that through so the UI can decide what to show.
func TestGetVersionHandlerWithoutLdflags(t *testing.T) {
	handler := &Handler{buildInfo: BuildInfo{Version: "dev"}}

	recorder := httptest.NewRecorder()
	handler.getVersionHandler(recorder, httptest.NewRequest(http.MethodGet, "/api/version", nil))

	var got BuildInfo
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.Version != "dev" {
		t.Errorf("version = %q, want %q", got.Version, "dev")
	}
}
