package application

import (
	"context"
	"regexp"
	"strings"
	"time"

	projectsdomain "github.com/devsvault/devsvault/apps/api/internal/projects/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
)

var (
	ErrNotFound     = projectsdomain.ErrNotFound
	ErrInvalidInput = projectsdomain.ErrInvalidInput
	ErrSlugTaken    = projectsdomain.ErrSlugTaken
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$`)

type Repository interface {
	Create(ctx context.Context, project projectsdomain.Project) error
	ListByWorkspace(ctx context.Context, workspaceID string) ([]projectsdomain.Project, error)
	FindByID(ctx context.Context, id string) (projectsdomain.Project, error)
	FindBySlug(ctx context.Context, workspaceID string, slug string) (projectsdomain.Project, error)
	Update(ctx context.Context, project projectsdomain.Project) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, workspaceID string, name string, slug string, description string, createdBy string) (projectsdomain.Project, error) {
	workspaceID, name, slug, description = strings.TrimSpace(workspaceID), strings.TrimSpace(name), strings.TrimSpace(slug), strings.TrimSpace(description)
	if workspaceID == "" || !validName(name) || !validSlug(slug) || len(description) > 300 || strings.TrimSpace(createdBy) == "" {
		return projectsdomain.Project{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	project := projectsdomain.Project{ID: shared.NewID("prj"), WorkspaceID: workspaceID, Name: name, Slug: slug, Description: description, CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, project); err != nil {
		return projectsdomain.Project{}, err
	}
	return project, nil
}

func (s *Service) List(ctx context.Context, workspaceID string) ([]projectsdomain.Project, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *Service) Get(ctx context.Context, id string) (projectsdomain.Project, error) {
	if strings.TrimSpace(id) == "" {
		return projectsdomain.Project{}, ErrInvalidInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, name string, description string) (projectsdomain.Project, error) {
	name, description = strings.TrimSpace(name), strings.TrimSpace(description)
	if strings.TrimSpace(id) == "" || !validName(name) || len(description) > 300 {
		return projectsdomain.Project{}, ErrInvalidInput
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return projectsdomain.Project{}, err
	}
	project.Name = name
	project.Description = description
	project.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, project); err != nil {
		return projectsdomain.Project{}, err
	}
	return project, nil
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
