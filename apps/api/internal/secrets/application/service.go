package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	encdomain "github.com/devsvault/devsvault/apps/api/internal/encryption/domain"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	secretsdomain "github.com/devsvault/devsvault/apps/api/internal/secrets/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
)

var (
	ErrInvalidInput = secretsdomain.ErrInvalidInput
	ErrNotFound     = secretsdomain.ErrNotFound
)

var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

type Encryptor interface {
	Encrypt(plaintext []byte, aad []byte) (encdomain.EncryptedPayload, error)
	Decrypt(payload encdomain.EncryptedPayload, aad []byte) ([]byte, error)
}

type PolicyChecker interface {
	Authorize(ctx context.Context, actor authdomain.Actor, action policiesapp.Action, resource policiesapp.Resource) error
}

type Repository interface {
	Create(ctx context.Context, secret secretsdomain.Secret, version secretsdomain.SecretVersion) error
	List(ctx context.Context) ([]secretsdomain.SecretMetadata, error)
	FindByID(ctx context.Context, id string) (secretsdomain.Secret, error)
	FindByPath(ctx context.Context, logicalPath string) (secretsdomain.Secret, error)
	ActiveVersion(ctx context.Context, secretID string) (secretsdomain.SecretVersion, error)
	AddVersion(ctx context.Context, secretID string, version secretsdomain.SecretVersion) error
	RevokeVersion(ctx context.Context, secretID string, version int, at time.Time) error
	MarkAccessed(ctx context.Context, secretID string, at time.Time) error
}

type Service struct {
	repo      Repository
	encryptor Encryptor
	policy    PolicyChecker
	audit     *auditapp.Service
}

type CreateInput struct {
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id"`
	EnvironmentID string `json:"environment_id"`
	Name          string `json:"name"`
	Value         string `json:"value"`
}

type RotateInput struct {
	SecretID string `json:"secret_id"`
	Value    string `json:"value"`
}

type ResolvedSecret struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
	Value   string `json:"value"`
}

func NewService(repo Repository, encryptor Encryptor, policy PolicyChecker, audit *auditapp.Service) *Service {
	return &Service{repo: repo, encryptor: encryptor, policy: policy, audit: audit}
}

func (s *Service) Create(ctx context.Context, actor authdomain.Actor, input CreateInput) (secretsdomain.SecretMetadata, error) {
	if err := validateCreate(input); err != nil {
		return secretsdomain.SecretMetadata{}, err
	}
	resource := policiesapp.Resource{WorkspaceID: input.WorkspaceID, ProjectID: input.ProjectID, EnvironmentID: input.EnvironmentID}
	if err := s.policy.Authorize(ctx, actor, policiesapp.ActionSecretWrite, resource); err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.create", ResourceType: "secret", Outcome: auditapp.OutcomeDenied})
		return secretsdomain.SecretMetadata{}, err
	}

	now := time.Now().UTC()
	secretID := shared.NewID("sec")
	logicalPath := buildPath(input.WorkspaceID, input.ProjectID, input.EnvironmentID, input.Name)
	payload, err := s.encryptor.Encrypt([]byte(input.Value), aad(secretID, 1, logicalPath))
	if err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.create", ResourceType: "secret", ResourceID: secretID, Outcome: auditapp.OutcomeError})
		return secretsdomain.SecretMetadata{}, err
	}
	secret := secretsdomain.Secret{
		ID:            secretID,
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		EnvironmentID: input.EnvironmentID,
		Name:          input.Name,
		LogicalPath:   logicalPath,
		ActiveVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	version := secretsdomain.SecretVersion{SecretID: secretID, Version: 1, Payload: payload, CreatedAt: now}
	if err := s.repo.Create(ctx, secret, version); err != nil {
		return secretsdomain.SecretMetadata{}, err
	}
	s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.create", ResourceType: "secret", ResourceID: secretID, Outcome: auditapp.OutcomeSuccess})
	return metadata(secret), nil
}

func (s *Service) List(ctx context.Context, actor authdomain.Actor) ([]secretsdomain.SecretMetadata, error) {
	if err := s.policy.Authorize(ctx, actor, policiesapp.ActionSecretListMetadata, policiesapp.Resource{}); err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.list", ResourceType: "secret", Outcome: auditapp.OutcomeDenied})
		return nil, err
	}
	return s.repo.List(ctx)
}

func (s *Service) Resolve(ctx context.Context, actor authdomain.Actor, logicalPath string) (ResolvedSecret, error) {
	logicalPath = strings.TrimSpace(logicalPath)
	if !validPath(logicalPath) {
		return ResolvedSecret{}, ErrInvalidInput
	}
	secret, err := s.repo.FindByPath(ctx, logicalPath)
	if err != nil {
		return ResolvedSecret{}, ErrNotFound
	}
	resource := policiesapp.Resource{WorkspaceID: secret.WorkspaceID, ProjectID: secret.ProjectID, EnvironmentID: secret.EnvironmentID, SecretID: secret.ID}
	if err := s.policy.Authorize(ctx, actor, policiesapp.ActionSecretReadValue, resource); err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.read", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeDenied})
		return ResolvedSecret{}, err
	}
	version, err := s.repo.ActiveVersion(ctx, secret.ID)
	if err != nil {
		return ResolvedSecret{}, ErrNotFound
	}
	plaintext, err := s.encryptor.Decrypt(version.Payload, aad(secret.ID, version.Version, secret.LogicalPath))
	if err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.read", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeError})
		return ResolvedSecret{}, err
	}
	accessedAt := time.Now().UTC()
	_ = s.repo.MarkAccessed(ctx, secret.ID, accessedAt)
	s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.read", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeSuccess})
	return ResolvedSecret{Path: secret.LogicalPath, Version: version.Version, Value: string(plaintext)}, nil
}

func (s *Service) Rotate(ctx context.Context, actor authdomain.Actor, input RotateInput) (secretsdomain.SecretMetadata, error) {
	if strings.TrimSpace(input.SecretID) == "" || input.Value == "" || len(input.Value) > 65536 {
		return secretsdomain.SecretMetadata{}, ErrInvalidInput
	}
	secret, err := s.repo.FindByID(ctx, input.SecretID)
	if err != nil {
		return secretsdomain.SecretMetadata{}, ErrNotFound
	}
	resource := policiesapp.Resource{WorkspaceID: secret.WorkspaceID, ProjectID: secret.ProjectID, EnvironmentID: secret.EnvironmentID, SecretID: secret.ID}
	if err := s.policy.Authorize(ctx, actor, policiesapp.ActionSecretRotate, resource); err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.rotate", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeDenied})
		return secretsdomain.SecretMetadata{}, err
	}
	nextVersion := secret.ActiveVersion + 1
	payload, err := s.encryptor.Encrypt([]byte(input.Value), aad(secret.ID, nextVersion, secret.LogicalPath))
	if err != nil {
		return secretsdomain.SecretMetadata{}, err
	}
	if err := s.repo.AddVersion(ctx, secret.ID, secretsdomain.SecretVersion{SecretID: secret.ID, Version: nextVersion, Payload: payload, CreatedAt: time.Now().UTC()}); err != nil {
		return secretsdomain.SecretMetadata{}, err
	}
	s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.rotate", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeSuccess})
	secret.ActiveVersion = nextVersion
	secret.UpdatedAt = time.Now().UTC()
	return metadata(secret), nil
}

func (s *Service) RevokeVersion(ctx context.Context, actor authdomain.Actor, secretID string, version int) error {
	secret, err := s.repo.FindByID(ctx, secretID)
	if err != nil {
		return ErrNotFound
	}
	resource := policiesapp.Resource{WorkspaceID: secret.WorkspaceID, ProjectID: secret.ProjectID, EnvironmentID: secret.EnvironmentID, SecretID: secret.ID}
	if err := s.policy.Authorize(ctx, actor, policiesapp.ActionSecretRevoke, resource); err != nil {
		s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.revoke", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeDenied})
		return err
	}
	if err := s.repo.RevokeVersion(ctx, secretID, version, time.Now().UTC()); err != nil {
		return err
	}
	s.audit.Record(ctx, auditapp.EventInput{Actor: actor, Action: "secret.revoke", ResourceType: "secret", ResourceID: secret.ID, Outcome: auditapp.OutcomeSuccess})
	return nil
}

func validateCreate(input CreateInput) error {
	if !segmentPattern.MatchString(input.WorkspaceID) || !segmentPattern.MatchString(input.ProjectID) || !segmentPattern.MatchString(input.EnvironmentID) || !segmentPattern.MatchString(input.Name) {
		return ErrInvalidInput
	}
	if input.Value == "" || len(input.Value) > 65536 {
		return ErrInvalidInput
	}
	return nil
}

func buildPath(workspaceID string, projectID string, environmentID string, name string) string {
	return workspaceID + "/" + projectID + "/" + environmentID + "/" + name
}

func validPath(path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if !segmentPattern.MatchString(part) {
			return false
		}
	}
	return true
}

func aad(secretID string, version int, logicalPath string) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s", secretID, version, logicalPath))
}

func metadata(secret secretsdomain.Secret) secretsdomain.SecretMetadata {
	return secretsdomain.SecretMetadata{
		ID:             secret.ID,
		WorkspaceID:    secret.WorkspaceID,
		ProjectID:      secret.ProjectID,
		EnvironmentID:  secret.EnvironmentID,
		Name:           secret.Name,
		LogicalPath:    secret.LogicalPath,
		ActiveVersion:  secret.ActiveVersion,
		CreatedAt:      secret.CreatedAt,
		UpdatedAt:      secret.UpdatedAt,
		LastAccessedAt: secret.LastAccessedAt,
	}
}
