package application

import (
	"context"
	"regexp"
	"strings"
	"time"

	environmentsdomain "github.com/devsvault/devsvault/apps/api/internal/environments/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
)

var (
	ErrNotFound     = environmentsdomain.ErrNotFound
	ErrInvalidInput = environmentsdomain.ErrInvalidInput
	ErrSlugTaken    = environmentsdomain.ErrSlugTaken
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$`)

type Repository interface {
	Create(ctx context.Context, environment environmentsdomain.Environment) error
	ListByProject(ctx context.Context, projectID string) ([]environmentsdomain.Environment, error)
	FindByID(ctx context.Context, id string) (environmentsdomain.Environment, error)
	FindBySlug(ctx context.Context, projectID string, slug string) (environmentsdomain.Environment, error)
	Delete(ctx context.Context, id string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, projectID string, name string, slug string, createdBy string) (environmentsdomain.Environment, error) {
	projectID, name, slug = strings.TrimSpace(projectID), strings.TrimSpace(name), strings.TrimSpace(slug)
	if projectID == "" || !validName(name) || !validSlug(slug) || strings.TrimSpace(createdBy) == "" {
		return environmentsdomain.Environment{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	environment := environmentsdomain.Environment{ID: shared.NewID("env"), ProjectID: projectID, Name: name, Slug: slug, CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, environment); err != nil {
		return environmentsdomain.Environment{}, err
	}
	return environment, nil
}

func (s *Service) List(ctx context.Context, projectID string) ([]environmentsdomain.Environment, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Get(ctx context.Context, id string) (environmentsdomain.Environment, error) {
	if strings.TrimSpace(id) == "" {
		return environmentsdomain.Environment{}, ErrInvalidInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func validName(name string) bool { return name != "" && len(name) <= 100 }

func validSlug(slug string) bool {
	return len(slug) >= 2 && len(slug) <= 50 && slugPattern.MatchString(slug)
}
