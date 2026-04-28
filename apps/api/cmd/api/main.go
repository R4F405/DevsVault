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
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	secretsapp "github.com/devsvault/devsvault/apps/api/internal/secrets/application"
	secretsinfra "github.com/devsvault/devsvault/apps/api/internal/secrets/infrastructure"
	httpapi "github.com/devsvault/devsvault/apps/api/internal/server/interfaces/http"
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

	auditRepo := auditinfra.NewMemoryRepository()
	auditService := auditapp.NewService(auditRepo)
	policyService := policiesapp.NewAuthorizer(policiesapp.DefaultRoleBindings())
	encryptionService := encapp.NewEnvelopeService(encinfra.NewStaticKEKProvider("dev-local-key", masterKey))
	secretRepo := secretsinfra.NewMemoryRepository()
	secretService := secretsapp.NewService(secretRepo, encryptionService, policyService, auditService)
	authService := authapp.NewService(authapp.NewHMACTokenIssuer(signingKey, time.Hour), auditService)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Auth:    authService,
		Secrets: secretService,
		Audit:   auditService,
		Policy:  policyService,
		Logger:  logger,
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
