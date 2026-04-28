package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devsvault/devsvault/apps/cli/internal/config"
)

func TestResourceCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspaces":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": "w1", "name": "Acme", "slug": "acme"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspaces":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "w1", "name": "Acme", "slug": "acme"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspaces/w1/projects":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": "p1", "workspace_id": "w1", "name": "API", "slug": "api"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspaces/w1/projects":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "p1", "workspace_id": "w1", "name": "API", "slug": "api"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/p1/environments":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": "e1", "project_id": "p1", "name": "Development", "slug": "dev"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/p1/environments":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "e1", "project_id": "p1", "name": "Development", "slug": "dev"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	withCommandTempHome(t)
	if err := config.Save(config.Session{APIURL: server.URL, AccessToken: "token-value", ExpiresAt: time.Now().UTC().Add(time.Hour)}); err != nil {
		t.Fatalf("save session failed: %v", err)
	}

	runResourceCommand(t, []string{"workspaces", "create", "Acme", "acme"}, "Workspace created: w1")
	runResourceCommand(t, []string{"workspaces", "list"}, "acme")
	runResourceCommand(t, []string{"projects", "create", "w1", "API", "api"}, "Project created: p1")
	runResourceCommand(t, []string{"projects", "list", "w1"}, "api")
	runResourceCommand(t, []string{"environments", "create", "p1", "Development", "dev"}, "Environment created: e1")
	runResourceCommand(t, []string{"environments", "list", "p1"}, "dev")
}

func runResourceCommand(t *testing.T, args []string, expected string) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("command %v failed: %v stderr=%s", args, err, stderr.String())
	}
	if !strings.Contains(stdout.String(), expected) {
		t.Fatalf("command %v expected %q in %q", args, expected, stdout.String())
	}
}
