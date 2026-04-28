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
