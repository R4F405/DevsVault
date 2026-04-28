package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devsvault/devsvault/apps/cli/internal/config"
)

func TestRunInjectsSecretsWithoutPrintingValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/secrets":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "sec-1", "name": "api-token", "logical_path": "w/p/e/api-token", "active_version": 1}}})
		case "/api/v1/secrets/resolve":
			json.NewEncoder(w).Encode(map[string]any{"value": "super-secret"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	withCommandTempHome(t)
	if err := config.Save(config.Session{APIURL: server.URL, AccessToken: "token-value", ExpiresAt: time.Now().UTC().Add(time.Hour)}); err != nil {
		t.Fatalf("save session failed: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"run", "--", os.Args[0], "-test.run=TestRunHelperProcess", "--"})
	t.Setenv("DEVSVAULT_HELPER_PROCESS", "1")
	verbose = false

	if err := root.Execute(); err != nil {
		t.Fatalf("run command failed: %v; stderr=%s", err, stderr.String())
	}
	combined := stdout.String() + stderr.String()
	if strings.Contains(combined, "super-secret") {
		t.Fatalf("secret value leaked in command output: %q", combined)
	}
	if !strings.Contains(stdout.String(), "helper-ok") {
		t.Fatalf("expected helper output, got %q", stdout.String())
	}
}

func TestRunHelperProcess(t *testing.T) {
	if os.Getenv("DEVSVAULT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("API_TOKEN") != "super-secret" {
		os.Exit(2)
	}
	_, _ = os.Stdout.WriteString("helper-ok")
	os.Exit(0)
}

func withCommandTempHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
}
