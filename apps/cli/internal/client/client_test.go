package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientLoginAndSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login":
			json.NewEncoder(w).Encode(map[string]any{"access_token": "token-value", "expires_in": 3600})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/secrets":
			if r.Header.Get("Authorization") != "Bearer token-value" {
				t.Fatalf("missing authorization header")
			}
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "sec-1", "name": "API_KEY", "logical_path": "w/p/e/API_KEY", "active_version": 2}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/secrets/resolve":
			json.NewEncoder(w).Encode(map[string]any{"value": "secret-value"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/secrets":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/secrets/sec-1/versions":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/secrets/sec-1/versions/2/revoke":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspaces":
			json.NewEncoder(w).Encode(map[string]any{"id": "w1", "name": "Acme", "slug": "acme", "description": "Main"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspaces":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "w1", "name": "Acme", "slug": "acme"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspaces/w1":
			json.NewEncoder(w).Encode(map[string]any{"id": "w1", "name": "Acme", "slug": "acme"})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/workspaces/w1":
			json.NewEncoder(w).Encode(map[string]any{"id": "w1", "name": "Acme Updated", "slug": "acme"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/workspaces/w1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspaces/w1/projects":
			json.NewEncoder(w).Encode(map[string]any{"id": "p1", "workspace_id": "w1", "name": "API", "slug": "api"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspaces/w1/projects":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "p1", "workspace_id": "w1", "name": "API", "slug": "api"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/p1":
			json.NewEncoder(w).Encode(map[string]any{"id": "p1", "workspace_id": "w1", "name": "API", "slug": "api"})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/workspaces/w1/projects/p1":
			json.NewEncoder(w).Encode(map[string]any{"id": "p1", "workspace_id": "w1", "name": "API Updated", "slug": "api"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/workspaces/w1/projects/p1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/p1/environments":
			json.NewEncoder(w).Encode(map[string]any{"id": "e1", "project_id": "p1", "name": "Development", "slug": "dev"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/p1/environments":
			json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"id": "e1", "project_id": "p1", "name": "Development", "slug": "dev"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/environments/e1":
			json.NewEncoder(w).Encode(map[string]any{"id": "e1", "project_id": "p1", "name": "Development", "slug": "dev"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/p1/environments/e1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	apiClient := New(server.URL, "")
	token, _, err := apiClient.Login("admin@example.local", "user")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token != "token-value" {
		t.Fatalf("unexpected token")
	}
	apiClient.Token = token

	items, err := apiClient.ListSecrets()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(items) != 1 || items[0].LogicalPath != "w/p/e/API_KEY" {
		t.Fatalf("unexpected items: %#v", items)
	}
	value, err := apiClient.GetSecret("w/p/e/API_KEY")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if value != "secret-value" {
		t.Fatalf("unexpected value")
	}
	if err := apiClient.CreateSecret("w", "p", "e", "API_KEY", "secret-value"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := apiClient.RotateSecret("sec-1", "new-secret-value"); err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	if err := apiClient.RevokeVersion("sec-1", 2); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	workspace, err := apiClient.CreateWorkspace("Acme", "acme", "Main")
	if err != nil || workspace.ID != "w1" {
		t.Fatalf("create workspace failed: workspace=%#v err=%v", workspace, err)
	}
	workspaces, err := apiClient.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("list workspaces failed: items=%#v err=%v", workspaces, err)
	}
	if _, err := apiClient.GetWorkspace("w1"); err != nil {
		t.Fatalf("get workspace failed: %v", err)
	}
	if _, err := apiClient.UpdateWorkspace("w1", "Acme Updated", "Updated"); err != nil {
		t.Fatalf("update workspace failed: %v", err)
	}
	project, err := apiClient.CreateProject("w1", "API", "api", "Backend")
	if err != nil || project.ID != "p1" {
		t.Fatalf("create project failed: project=%#v err=%v", project, err)
	}
	projects, err := apiClient.ListProjects("w1")
	if err != nil || len(projects) != 1 {
		t.Fatalf("list projects failed: items=%#v err=%v", projects, err)
	}
	if _, err := apiClient.GetProject("p1"); err != nil {
		t.Fatalf("get project failed: %v", err)
	}
	if _, err := apiClient.UpdateProject("w1", "p1", "API Updated", "Updated"); err != nil {
		t.Fatalf("update project failed: %v", err)
	}
	environment, err := apiClient.CreateEnvironment("p1", "Development", "dev")
	if err != nil || environment.ID != "e1" {
		t.Fatalf("create environment failed: environment=%#v err=%v", environment, err)
	}
	environments, err := apiClient.ListEnvironments("p1")
	if err != nil || len(environments) != 1 {
		t.Fatalf("list environments failed: items=%#v err=%v", environments, err)
	}
	if _, err := apiClient.GetEnvironment("e1"); err != nil {
		t.Fatalf("get environment failed: %v", err)
	}
	if err := apiClient.DeleteEnvironment("p1", "e1"); err != nil {
		t.Fatalf("delete environment failed: %v", err)
	}
	if err := apiClient.DeleteProject("w1", "p1"); err != nil {
		t.Fatalf("delete project failed: %v", err)
	}
	if err := apiClient.DeleteWorkspace("w1"); err != nil {
		t.Fatalf("delete workspace failed: %v", err)
	}
}

func TestClientErrorDoesNotExposeResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "secret-value token-value", http.StatusForbidden)
	}))
	defer server.Close()

	apiClient := New(server.URL, "token-value")
	err := apiClient.RotateSecret("sec-1", "secret-value")
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	if strings.Contains(message, "secret-value") || strings.Contains(message, "token-value") {
		t.Fatalf("error exposed sensitive data: %q", message)
	}
	if !strings.Contains(message, "status 403") {
		t.Fatalf("expected status in error: %q", message)
	}
}
