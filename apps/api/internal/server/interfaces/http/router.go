package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	authapp "github.com/devsvault/devsvault/apps/api/internal/auth/application"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	secretsapp "github.com/devsvault/devsvault/apps/api/internal/secrets/application"
)

type Dependencies struct {
	Auth    *authapp.Service
	Secrets *secretsapp.Service
	Audit   *auditapp.Service
	Policy  *policiesapp.Authorizer
	Logger  *slog.Logger
}

type router struct {
	deps Dependencies
}

type contextKey string

const actorKey contextKey = "actor"

func NewRouter(deps Dependencies) http.Handler {
	r := &router{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", r.health)
	mux.HandleFunc("POST /api/v1/auth/login", r.login)
	mux.Handle("GET /api/v1/secrets", r.withAuth(http.HandlerFunc(r.listSecrets)))
	mux.Handle("POST /api/v1/secrets", r.withAuth(http.HandlerFunc(r.createSecret)))
	mux.Handle("GET /api/v1/secrets/resolve", r.withAuth(http.HandlerFunc(r.resolveSecret)))
	mux.Handle("POST /api/v1/secrets/", r.withAuth(http.HandlerFunc(r.secretAction)))
	mux.Handle("GET /api/v1/audit/events", r.withAuth(http.HandlerFunc(r.listAudit)))
	return securityHeaders(mux)
}

func (r *router) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *router) login(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Subject   string `json:"subject"`
		ActorType string `json:"actor_type"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	token, err := r.deps.Auth.Login(req.Context(), authapp.LoginInput{Subject: input.Subject, ActorType: authdomain.ActorType(input.ActorType)})
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token.AccessToken, "token_type": "Bearer", "expires_in": int(time.Until(token.ExpiresAt).Seconds())})
}

func (r *router) listSecrets(w http.ResponseWriter, req *http.Request) {
	items, err := r.deps.Secrets.List(req.Context(), actorFrom(req.Context()))
	if err != nil {
		writeError(w, statusFromError(err), "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *router) createSecret(w http.ResponseWriter, req *http.Request) {
	var input secretsapp.CreateInput
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	created, err := r.deps.Secrets.Create(req.Context(), actorFrom(req.Context()), input)
	if err != nil {
		writeError(w, statusFromError(err), "secret could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *router) resolveSecret(w http.ResponseWriter, req *http.Request) {
	resolved, err := r.deps.Secrets.Resolve(req.Context(), actorFrom(req.Context()), req.URL.Query().Get("path"))
	if err != nil {
		writeError(w, statusFromError(err), "secret could not be resolved")
		return
	}
	writeJSON(w, http.StatusOK, resolved)
}

func (r *router) secretAction(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/v1/secrets/"), "/")
	if len(parts) == 2 && parts[1] == "versions" {
		var body struct {
			Value string `json:"value"`
		}
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")
			return
		}
		updated, err := r.deps.Secrets.Rotate(req.Context(), actorFrom(req.Context()), secretsapp.RotateInput{SecretID: parts[0], Value: body.Value})
		if err != nil {
			writeError(w, statusFromError(err), "secret could not be rotated")
			return
		}
		writeJSON(w, http.StatusOK, updated)
		return
	}
	if len(parts) == 4 && parts[1] == "versions" && parts[3] == "revoke" {
		version, err := strconv.Atoi(parts[2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := r.deps.Secrets.RevokeVersion(req.Context(), actorFrom(req.Context()), parts[0], version); err != nil {
			writeError(w, statusFromError(err), "secret version could not be revoked")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

func (r *router) listAudit(w http.ResponseWriter, req *http.Request) {
	actor := actorFrom(req.Context())
	if err := r.deps.Policy.Authorize(req.Context(), actor, policiesapp.ActionAuditRead, policiesapp.Resource{}); err != nil {
		r.deps.Audit.Record(req.Context(), auditapp.EventInput{Actor: actor, Action: "audit.read", ResourceType: "audit_log", Outcome: auditapp.OutcomeDenied})
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	items, err := r.deps.Audit.List(req.Context(), 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "audit events unavailable")
		return
	}
	r.deps.Audit.Record(req.Context(), auditapp.EventInput{Actor: actor, Action: "audit.read", ResourceType: "audit_log", Outcome: auditapp.OutcomeSuccess})
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *router) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		actor, err := r.deps.Auth.Authenticate(req.Context(), token)
		if err != nil {
			r.deps.Audit.Record(req.Context(), auditapp.EventInput{Actor: authdomain.Anonymous(), Action: "auth.authorize", ResourceType: "request", Outcome: auditapp.OutcomeDenied})
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), actorKey, actor)))
	})
}

func actorFrom(ctx context.Context) authdomain.Actor {
	actor, ok := ctx.Value(actorKey).(authdomain.Actor)
	if !ok {
		return authdomain.Anonymous()
	}
	return actor
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, req)
	})
}

func decodeJSON(req *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, req.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, policiesapp.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, secretsapp.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, secretsapp.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
