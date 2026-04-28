package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	auditinfra "github.com/devsvault/devsvault/apps/api/internal/audit/infrastructure"
	authapp "github.com/devsvault/devsvault/apps/api/internal/auth/application"
	encapp "github.com/devsvault/devsvault/apps/api/internal/encryption/application"
	encinfra "github.com/devsvault/devsvault/apps/api/internal/encryption/infrastructure"
	environmentsapp "github.com/devsvault/devsvault/apps/api/internal/environments/application"
	environmentsinfra "github.com/devsvault/devsvault/apps/api/internal/environments/infrastructure"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	policiesinfra "github.com/devsvault/devsvault/apps/api/internal/policies/infrastructure"
	projectsapp "github.com/devsvault/devsvault/apps/api/internal/projects/application"
	projectsinfra "github.com/devsvault/devsvault/apps/api/internal/projects/infrastructure"
	secretsapp "github.com/devsvault/devsvault/apps/api/internal/secrets/application"
	secretsinfra "github.com/devsvault/devsvault/apps/api/internal/secrets/infrastructure"
	httpapi "github.com/devsvault/devsvault/apps/api/internal/server/interfaces/http"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
	workspacesapp "github.com/devsvault/devsvault/apps/api/internal/workspaces/application"
	workspacesinfra "github.com/devsvault/devsvault/apps/api/internal/workspaces/infrastructure"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	addr := getenv("DEVSVAULT_API_ADDR", ":8080")

	masterKey, err := loadBase64Key("DEVSVAULT_MASTER_KEY_B64")
	if err != nil {
		logger.Error("missing development master key", "error", err)
		os.Exit(1)
	}

	signingKey, err := loadBase64Key("DEVSVAULT_AUTH_SIGNING_KEY")
	if err != nil {
		logger.Error("missing development auth signing key", "error", err)
		os.Exit(1)
	}

	var auditRepo auditapp.Repository
	var secretRepo secretsapp.Repository
	var workspaceRepo workspacesapp.Repository
	var projectRepo projectsapp.Repository
	var environmentRepo environmentsapp.Repository
	var policyService *policiesapp.Authorizer

	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		pool, err := postgres.NewPool(context.Background(), databaseURL)
		if err != nil {
			logger.Error("postgres connection failed", "error", "database unavailable")
			os.Exit(1)
		}
		defer pool.Close()
		logger.Info("using postgres repository")
		auditRepo = auditinfra.NewPostgresRepository(pool)
		secretRepo = secretsinfra.NewPostgresRepository(pool)
		workspaceRepo = workspacesinfra.NewPostgresRepository(pool)
		projectRepo = projectsinfra.NewPostgresRepository(pool)
		environmentRepo = environmentsinfra.NewPostgresRepository(pool)
		policyService = policiesapp.NewAuthorizerWithStore(policiesapp.DefaultRoleBindings(), policiesinfra.NewPostgresRepository(pool))
	} else {
		logger.Info("using in-memory repository")
		auditRepo = auditinfra.NewMemoryRepository()
		secretRepo = secretsinfra.NewMemoryRepository()
		workspaceRepo = workspacesinfra.NewMemoryRepository()
		projectRepo = projectsinfra.NewMemoryRepository()
		environmentRepo = environmentsinfra.NewMemoryRepository()
		policyService = policiesapp.NewAuthorizer(policiesapp.DefaultRoleBindings())
	}

	auditService := auditapp.NewService(auditRepo)
	encryptionService := encapp.NewEnvelopeService(encinfra.NewStaticKEKProvider("dev-local-key", masterKey))
	secretService := secretsapp.NewService(secretRepo, encryptionService, policyService, auditService)
	authService := authapp.NewService(authapp.NewHMACTokenIssuer(signingKey, time.Hour), auditService)
	workspaceService := workspacesapp.NewService(workspaceRepo)
	projectService := projectsapp.NewService(projectRepo)
	environmentService := environmentsapp.NewService(environmentRepo)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Auth:         authService,
		Secrets:      secretService,
		Audit:        auditService,
		Policy:       policyService,
		Workspaces:   workspaceService,
		Projects:     projectService,
		Environments: environmentService,
		Logger:       logger,
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadBase64Key(name string) ([]byte, error) {
	value := os.Getenv(name)
	if value == "" {
		return nil, errors.New("required environment variable is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, errors.New("required environment variable must be base64")
	}
	if len(decoded) != 32 {
		return nil, errors.New("required environment variable must decode to 32 bytes")
	}
	return decoded, nil
}
