package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	auditinfra "github.com/devsvault/devsvault/apps/api/internal/audit/infrastructure"
	authapp "github.com/devsvault/devsvault/apps/api/internal/auth/application"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	environmentsapp "github.com/devsvault/devsvault/apps/api/internal/environments/application"
	environmentsinfra "github.com/devsvault/devsvault/apps/api/internal/environments/infrastructure"
	projectsapp "github.com/devsvault/devsvault/apps/api/internal/projects/application"
	projectsinfra "github.com/devsvault/devsvault/apps/api/internal/projects/infrastructure"
	workspacesapp "github.com/devsvault/devsvault/apps/api/internal/workspaces/application"
	workspacesinfra "github.com/devsvault/devsvault/apps/api/internal/workspaces/infrastructure"
)

func TestResourceRoutesRequireAuthentication(t *testing.T) {
	handler, _ := testResourceRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rec.Code)
	}
}

func TestResourceRoutesLifecycle(t *testing.T) {
	handler, token := testResourceRouter(t)

	workspace := doResourceRequest(t, handler, token, http.MethodPost, "/api/v1/workspaces", map[string]string{"name": "Acme", "slug": "acme", "description": "Main"}, http.StatusCreated)
	workspaceID := workspace["id"].(string)

	list := doResourceRequest(t, handler, token, http.MethodGet, "/api/v1/workspaces", nil, http.StatusOK)
	if len(list["items"].([]any)) != 1 {
		t.Fatalf("expected workspace in list: %#v", list)
	}

	project := doResourceRequest(t, handler, token, http.MethodPost, "/api/v1/workspaces/"+workspaceID+"/projects", map[string]string{"name": "API", "slug": "api", "description": "Backend"}, http.StatusCreated)
	projectID := project["id"].(string)

	environment := doResourceRequest(t, handler, token, http.MethodPost, "/api/v1/projects/"+projectID+"/environments", map[string]string{"name": "Development", "slug": "dev"}, http.StatusCreated)
	environmentID := environment["id"].(string)

	gotProject := doResourceRequest(t, handler, token, http.MethodGet, "/api/v1/projects/"+projectID, nil, http.StatusOK)
	if gotProject["workspace_id"] != workspaceID {
		t.Fatalf("unexpected project: %#v", gotProject)
	}
	gotEnvironment := doResourceRequest(t, handler, token, http.MethodGet, "/api/v1/environments/"+environmentID, nil, http.StatusOK)
	if gotEnvironment["project_id"] != projectID {
		t.Fatalf("unexpected environment: %#v", gotEnvironment)
	}

	doResourceRequest(t, handler, token, http.MethodDelete, "/api/v1/projects/"+projectID+"/environments/"+environmentID, nil, http.StatusNoContent)
	doResourceRequest(t, handler, token, http.MethodDelete, "/api/v1/workspaces/"+workspaceID+"/projects/"+projectID, nil, http.StatusNoContent)
	doResourceRequest(t, handler, token, http.MethodDelete, "/api/v1/workspaces/"+workspaceID, nil, http.StatusNoContent)
}

func TestResourceRouteConflict(t *testing.T) {
	handler, token := testResourceRouter(t)
	doResourceRequest(t, handler, token, http.MethodPost, "/api/v1/workspaces", map[string]string{"name": "Acme", "slug": "acme"}, http.StatusCreated)
	doResourceRequest(t, handler, token, http.MethodPost, "/api/v1/workspaces", map[string]string{"name": "Acme 2", "slug": "acme"}, http.StatusConflict)
}

func testResourceRouter(t *testing.T) (http.Handler, string) {
	t.Helper()
	auditService := auditapp.NewService(auditinfra.NewMemoryRepository())
	authService := authapp.NewService(authapp.NewHMACTokenIssuer([]byte("01234567890123456789012345678901"), time.Hour), auditService)
	token, err := authService.Login(context.Background(), authapp.LoginInput{Subject: "admin@example.local", ActorType: authdomain.ActorUser})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	handler := NewRouter(Dependencies{
		Auth:         authService,
		Audit:        auditService,
		Workspaces:   workspacesapp.NewService(workspacesinfra.NewMemoryRepository()),
		Projects:     projectsapp.NewService(projectsinfra.NewMemoryRepository()),
		Environments: environmentsapp.NewService(environmentsinfra.NewMemoryRepository()),
	})
	return handler, token.AccessToken
}

func doResourceRequest(t *testing.T, handler http.Handler, token string, method string, path string, body any, status int) map[string]any {
	t.Helper()
	payload := bytes.NewReader(nil)
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s expected %d, got %d body=%s", method, path, status, rec.Code, rec.Body.String())
	}
	if status == http.StatusNoContent {
		return nil
	}
	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	return response
}
