package application

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/devsvault/devsvault/apps/api/internal/shared"
	workspacesdomain "github.com/devsvault/devsvault/apps/api/internal/workspaces/domain"
)

var (
	ErrNotFound     = workspacesdomain.ErrNotFound
	ErrInvalidInput = workspacesdomain.ErrInvalidInput
	ErrSlugTaken    = workspacesdomain.ErrSlugTaken
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$`)

type Repository interface {
	Create(ctx context.Context, workspace workspacesdomain.Workspace) error
	List(ctx context.Context) ([]workspacesdomain.Workspace, error)
	FindByID(ctx context.Context, id string) (workspacesdomain.Workspace, error)
	FindBySlug(ctx context.Context, slug string) (workspacesdomain.Workspace, error)
	Update(ctx context.Context, workspace workspacesdomain.Workspace) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name string, slug string, description string, createdBy string) (workspacesdomain.Workspace, error) {
	name, slug, description = strings.TrimSpace(name), strings.TrimSpace(slug), strings.TrimSpace(description)
	if !validName(name) || !validSlug(slug) || len(description) > 300 || strings.TrimSpace(createdBy) == "" {
		return workspacesdomain.Workspace{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	workspace := workspacesdomain.Workspace{ID: shared.NewID("wks"), Name: name, Slug: slug, Description: description, CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, workspace); err != nil {
		return workspacesdomain.Workspace{}, err
	}
	return workspace, nil
}

func (s *Service) List(ctx context.Context) ([]workspacesdomain.Workspace, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (workspacesdomain.Workspace, error) {
	if strings.TrimSpace(id) == "" {
		return workspacesdomain.Workspace{}, ErrInvalidInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, name string, description string) (workspacesdomain.Workspace, error) {
	name, description = strings.TrimSpace(name), strings.TrimSpace(description)
	if strings.TrimSpace(id) == "" || !validName(name) || len(description) > 300 {
		return workspacesdomain.Workspace{}, ErrInvalidInput
	}
	workspace, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return workspacesdomain.Workspace{}, err
	}
	workspace.Name = name
	workspace.Description = description
	workspace.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, workspace); err != nil {
		return workspacesdomain.Workspace{}, err
	}
	return workspace, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func validName(name string) bool {
	return name != "" && len(name) <= 100
}

func validSlug(slug string) bool {
	return len(slug) >= 2 && len(slug) <= 50 && slugPattern.MatchString(slug)
}
