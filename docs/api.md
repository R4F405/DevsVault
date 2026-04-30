# API actual

Base path: `/api/v1`.

Los endpoints, salvo `GET /healthz` y `POST /api/v1/auth/login`, requieren `Authorization: Bearer <token>`.

## Auth

### `POST /api/v1/auth/login`

Body:

```json
{"subject":"admin@example.local","actor_type":"user"}
```

Respuesta:

```json
{"access_token":"...","token_type":"Bearer","expires_in":3600}
```

Este login es solo un issuer de desarrollo. La interfaz esta preparada para reemplazarlo por OIDC.

## Workspaces

### `GET /api/v1/workspaces`

Lista workspaces.

### `POST /api/v1/workspaces`

Crea un workspace.

```json
{"name":"Local","slug":"local","description":"Local development"}
```

### `GET /api/v1/workspaces/{id}`

Devuelve un workspace por ID.

### `PATCH /api/v1/workspaces/{id}`

Actualiza nombre y descripcion.

```json
{"name":"Local","description":"Updated description"}
```

### `DELETE /api/v1/workspaces/{id}`

Elimina un workspace si no tiene dependencias bloqueantes.

## Projects

### `GET /api/v1/workspaces/{workspaceId}/projects`

Lista proyectos de un workspace.

### `POST /api/v1/workspaces/{workspaceId}/projects`

Crea un proyecto.

```json
{"name":"API","slug":"api","description":"Core API"}
```

### `GET /api/v1/workspaces/{workspaceId}/projects/{id}`

Devuelve un proyecto dentro de un workspace.

### `GET /api/v1/projects/{id}`

Devuelve un proyecto por ID global.

### `PATCH /api/v1/workspaces/{workspaceId}/projects/{id}`

Actualiza nombre y descripcion.

### `DELETE /api/v1/workspaces/{workspaceId}/projects/{id}`

Elimina un proyecto si no tiene dependencias bloqueantes.

## Environments

### `GET /api/v1/projects/{projectId}/environments`

Lista entornos de un proyecto.

### `POST /api/v1/projects/{projectId}/environments`

Crea un entorno.

```json
{"name":"Development","slug":"dev"}
```

### `GET /api/v1/projects/{projectId}/environments/{id}`

Devuelve un entorno dentro de un proyecto.

### `GET /api/v1/environments/{id}`

Devuelve un entorno por ID global.

### `DELETE /api/v1/projects/{projectId}/environments/{id}`

Elimina un entorno si no tiene dependencias bloqueantes.

## Secrets

### `GET /api/v1/secrets`

Lista metadatos. No devuelve valores completos.

### `POST /api/v1/secrets`

Crea secreto y version inicial.

```json
{
  "workspace_id":"workspace-local",
  "project_id":"project-api",
  "environment_id":"dev",
  "name":"DATABASE_URL",
  "value":"sensitive value"
}
```

### `GET /api/v1/secrets/resolve?path=workspace/project/env/name`

Devuelve el valor completo solo si el actor tiene `secret:read_value`.

### `POST /api/v1/secrets/{id}/versions`

Crea una nueva version activa.

### `POST /api/v1/secrets/{id}/versions/{version}/revoke`

Revoca una version.

## Audit

### `GET /api/v1/audit/events`

Devuelve eventos sanitizados para actores con `audit:read`.

## Errores

Los errores usan mensajes genericos, evitan detalles internos y nunca incluyen valores sensibles. Las validaciones estrictas rechazan JSON con campos desconocidos.