package http

import "net/http"

func (r *router) createWorkspace(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	created, err := r.deps.Workspaces.Create(req.Context(), input.Name, input.Slug, input.Description, actorFrom(req.Context()).ID)
	if err != nil {
		writeError(w, statusFromError(err), "workspace could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *router) listWorkspaces(w http.ResponseWriter, req *http.Request) {
	items, err := r.deps.Workspaces.List(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "workspaces unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *router) getWorkspace(w http.ResponseWriter, req *http.Request) {
	workspace, err := r.deps.Workspaces.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "workspace not found")
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (r *router) updateWorkspace(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	workspace, err := r.deps.Workspaces.Update(req.Context(), req.PathValue("id"), input.Name, input.Description)
	if err != nil {
		writeError(w, statusFromError(err), "workspace could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (r *router) deleteWorkspace(w http.ResponseWriter, req *http.Request) {
	if err := r.deps.Workspaces.Delete(req.Context(), req.PathValue("id")); err != nil {
		writeError(w, statusFromError(err), "workspace could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *router) createProject(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	created, err := r.deps.Projects.Create(req.Context(), req.PathValue("workspaceId"), input.Name, input.Slug, input.Description, actorFrom(req.Context()).ID)
	if err != nil {
		writeError(w, statusFromError(err), "project could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *router) listProjects(w http.ResponseWriter, req *http.Request) {
	items, err := r.deps.Projects.List(req.Context(), req.PathValue("workspaceId"))
	if err != nil {
		writeError(w, statusFromError(err), "projects unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *router) getProject(w http.ResponseWriter, req *http.Request) {
	project, err := r.deps.Projects.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "project not found")
		return
	}
	if project.WorkspaceID != req.PathValue("workspaceId") {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (r *router) getProjectByID(w http.ResponseWriter, req *http.Request) {
	project, err := r.deps.Projects.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "project not found")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (r *router) updateProject(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	existing, err := r.deps.Projects.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "project could not be updated")
		return
	}
	if existing.WorkspaceID != req.PathValue("workspaceId") {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	project, err := r.deps.Projects.Update(req.Context(), req.PathValue("id"), input.Name, input.Description)
	if err != nil {
		writeError(w, statusFromError(err), "project could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (r *router) deleteProject(w http.ResponseWriter, req *http.Request) {
	project, err := r.deps.Projects.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "project could not be deleted")
		return
	}
	if project.WorkspaceID != req.PathValue("workspaceId") {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err := r.deps.Projects.Delete(req.Context(), req.PathValue("id")); err != nil {
		writeError(w, statusFromError(err), "project could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *router) createEnvironment(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := decodeJSON(req, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	created, err := r.deps.Environments.Create(req.Context(), req.PathValue("projectId"), input.Name, input.Slug, actorFrom(req.Context()).ID)
	if err != nil {
		writeError(w, statusFromError(err), "environment could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *router) listEnvironments(w http.ResponseWriter, req *http.Request) {
	items, err := r.deps.Environments.List(req.Context(), req.PathValue("projectId"))
	if err != nil {
		writeError(w, statusFromError(err), "environments unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *router) getEnvironment(w http.ResponseWriter, req *http.Request) {
	environment, err := r.deps.Environments.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "environment not found")
		return
	}
	if environment.ProjectID != req.PathValue("projectId") {
		writeError(w, http.StatusNotFound, "environment not found")
		return
	}
	writeJSON(w, http.StatusOK, environment)
}

func (r *router) getEnvironmentByID(w http.ResponseWriter, req *http.Request) {
	environment, err := r.deps.Environments.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "environment not found")
		return
	}
	writeJSON(w, http.StatusOK, environment)
}

func (r *router) deleteEnvironment(w http.ResponseWriter, req *http.Request) {
	environment, err := r.deps.Environments.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), "environment could not be deleted")
		return
	}
	if environment.ProjectID != req.PathValue("projectId") {
		writeError(w, http.StatusNotFound, "environment not found")
		return
	}
	if err := r.deps.Environments.Delete(req.Context(), req.PathValue("id")); err != nil {
		writeError(w, statusFromError(err), "environment could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
