# API inicial

Base path: `/api/v1`.

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

Los errores evitan detalles internos y nunca incluyen valores sensibles.