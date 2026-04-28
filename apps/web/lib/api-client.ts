export type ActorType = "user" | "service";

export type LoginResponse = {
  access_token: string;
  token_type: string;
  expires_in: number;
};

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
};

export type Project = {
  id: string;
  workspace_id: string;
  name: string;
  slug: string;
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
};

export type Environment = {
  id: string;
  project_id: string;
  name: string;
  slug: string;
  created_by: string;
  created_at: string;
  updated_at: string;
};

export type SecretMetadata = {
  id: string;
  workspace_id: string;
  project_id: string;
  environment_id: string;
  name: string;
  logical_path: string;
  active_version: number;
  created_at: string;
  updated_at: string;
  last_accessed_at?: string | null;
};

export type AuditEvent = {
  id: string;
  occurred_at: string;
  actor_type: string;
  actor_id: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  outcome: "success" | "denied" | "error";
  metadata?: Record<string, string>;
};

type ListResponse<T> = { items: T[] };

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly endpoint: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly token?: string;

  constructor(baseUrl: string, token?: string) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.token = token;
  }

  login(input: { subject: string; actor_type: ActorType }) {
    return this.request<LoginResponse>("/api/v1/auth/login", { method: "POST", body: input, auth: false });
  }

  listWorkspaces() {
    return this.request<ListResponse<Workspace>>("/api/v1/workspaces").then((response) => response.items);
  }

  createWorkspace(input: { name: string; slug: string; description: string }) {
    return this.request<Workspace>("/api/v1/workspaces", { method: "POST", body: input });
  }

  getWorkspace(id: string) {
    return this.request<Workspace>(`/api/v1/workspaces/${encodeURIComponent(id)}`);
  }

  updateWorkspace(id: string, input: { name: string; description: string }) {
    return this.request<Workspace>(`/api/v1/workspaces/${encodeURIComponent(id)}`, { method: "PATCH", body: input });
  }

  deleteWorkspace(id: string) {
    return this.request<void>(`/api/v1/workspaces/${encodeURIComponent(id)}`, { method: "DELETE" });
  }

  listProjects(workspaceId: string) {
    return this.request<ListResponse<Project>>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects`).then((response) => response.items);
  }

  createProject(workspaceId: string, input: { name: string; slug: string; description: string }) {
    return this.request<Project>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects`, { method: "POST", body: input });
  }

  getProject(workspaceId: string, projectId: string) {
    return this.request<Project>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects/${encodeURIComponent(projectId)}`);
  }

  updateProject(workspaceId: string, projectId: string, input: { name: string; description: string }) {
    return this.request<Project>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects/${encodeURIComponent(projectId)}`, { method: "PATCH", body: input });
  }

  deleteProject(workspaceId: string, projectId: string) {
    return this.request<void>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects/${encodeURIComponent(projectId)}`, { method: "DELETE" });
  }

  listEnvironments(projectId: string) {
    return this.request<ListResponse<Environment>>(`/api/v1/projects/${encodeURIComponent(projectId)}/environments`).then((response) => response.items);
  }

  createEnvironment(projectId: string, input: { name: string; slug: string }) {
    return this.request<Environment>(`/api/v1/projects/${encodeURIComponent(projectId)}/environments`, { method: "POST", body: input });
  }

  deleteEnvironment(projectId: string, environmentId: string) {
    return this.request<void>(`/api/v1/projects/${encodeURIComponent(projectId)}/environments/${encodeURIComponent(environmentId)}`, { method: "DELETE" });
  }

  listSecrets() {
    return this.request<ListResponse<SecretMetadata>>("/api/v1/secrets").then((response) => response.items);
  }

  createSecret(input: { workspace_id: string; project_id: string; environment_id: string; name: string; value: string }) {
    return this.request<SecretMetadata>("/api/v1/secrets", { method: "POST", body: input });
  }

  resolveSecret(path: string) {
    return this.request<{ value: string }>(`/api/v1/secrets/resolve?path=${encodeURIComponent(path)}`);
  }

  rotateSecret(id: string, value: string) {
    return this.request<SecretMetadata>(`/api/v1/secrets/${encodeURIComponent(id)}/versions`, { method: "POST", body: { value } });
  }

  revokeSecretVersion(id: string, version: number) {
    return this.request<void>(`/api/v1/secrets/${encodeURIComponent(id)}/versions/${version}/revoke`, { method: "POST" });
  }

  listAuditEvents() {
    return this.request<ListResponse<AuditEvent>>("/api/v1/audit/events").then((response) => response.items);
  }

  private async request<T>(endpoint: string, options: { method?: string; body?: unknown; auth?: boolean } = {}): Promise<T> {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 10_000);
    const requiresAuth = options.auth ?? true;
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: options.method ?? "GET",
        signal: controller.signal,
        headers: {
          ...(options.body === undefined ? {} : { "Content-Type": "application/json" }),
          ...(requiresAuth && this.token ? { Authorization: `Bearer ${this.token}` } : {})
        },
        body: options.body === undefined ? undefined : JSON.stringify(options.body)
      });

      if (!response.ok) {
        throw new ApiError(`Request failed with status ${response.status}`, response.status, endpoint);
      }
      if (response.status === 204) {
        return undefined as T;
      }
      return (await response.json()) as T;
    } finally {
      window.clearTimeout(timeout);
    }
  }
}