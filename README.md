# DevsVault Secrets Manager

MVP de un secrets manager para desarrollo local y produccion. El sistema esta pensado para almacenar secretos estaticos versionados, recuperarlos por path logico, aplicar permisos por recurso y auditar acciones sensibles sin depender de archivos `.env` para valores sensibles.

## Arranque rapido

Requisitos minimos para levantar todo rapido:

- Docker Desktop
- Go 1.25+ si vas a ejecutar la API o los tests Go fuera de Docker
- Node.js 20+ si vas a ejecutar el panel fuera de Docker

Levanta PostgreSQL, API y panel web con Docker Compose:

```powershell
.\scripts\dev.ps1 -Build
```

Cuando termine, abre:

- Panel web: `http://localhost:3000`
- API: `http://localhost:8080`
- Healthcheck: `http://localhost:8080/healthz`

Login de desarrollo para el panel:

- API URL: `http://localhost:8080`
- Subject: `admin@example.local`
- Actor Type: `user`

CLI local contra la API levantada por Docker:

```powershell
go run ./apps/cli/cmd/devsvault login --url http://localhost:8080 --subject admin@example.local --type user
go run ./apps/cli/cmd/devsvault workspaces list
```

Para apagar el stack:

```powershell
docker compose down
```

Para borrar tambien los datos locales de PostgreSQL:

```powershell
docker compose down -v
```

## Estado del MVP

- Fase 1: arquitectura, modelo de datos, flujo de cifrado y permisos documentados.
- Fase 2: backend Go ejecutable con cifrado envelope AES-256-GCM, repositorios PostgreSQL y fallback en memoria.
- Fase 3: auth preparada para OIDC con tokens firmados de desarrollo y autorizacion centralizada.
- Fase 4: audit log sanitizado para lecturas, escrituras, revocaciones, logins y denegaciones.
- Fase 5: CLI Go con login/logout, CRUD de workspaces/proyectos/entornos, gestion de secretos y `run` con inyeccion efimera.
- Fase 6: panel Next.js + TypeScript conectado a la API para login, recursos, secretos y auditoria.
- Fase 7: SDK basico para Node y PHP/Laravel.

## Arquitectura

El monorepo separa aplicaciones, paquetes y documentacion:

```text
apps/api      Core API en Go
apps/cli      CLI en Go
apps/web      Panel admin en Next.js + TypeScript
packages      SDKs Node y PHP/Laravel
db            Migraciones SQL
docs          Arquitectura, API, cifrado y fases
```

El backend usa capas por modulo: `domain`, `application`, `infrastructure` e `interfaces`. Los modulos actuales son `auth`, `policies`, `secrets`, `encryption`, `audit`, `workspaces`, `projects` y `environments`.

## Setup local

Requisitos:

- Go 1.25+ para API y workspace Go completo
- Go 1.22+ para compilar solo la CLI con `GOWORK=off`
- Node.js 20+
- Docker Desktop

La forma recomendada para desarrollo completo es Docker Compose:

```powershell
.\scripts\dev.ps1 -Build
```

El compose usa PostgreSQL y aplica las migraciones montadas desde `db/migrations` al inicializar el volumen.

Si prefieres ejecutar piezas sueltas, levanta solo PostgreSQL:

```powershell
docker compose up -d postgres
```

Genera una master key de desarrollo fuera del repo y exportala en la terminal:

```powershell
$bytes = New-Object byte[] 32
$rng = New-Object System.Security.Cryptography.RNGCryptoServiceProvider
$rng.GetBytes($bytes)
$rng.Dispose()
$key = [Convert]::ToBase64String($bytes)
$env:DEVSVAULT_MASTER_KEY_B64 = $key
$env:DEVSVAULT_AUTH_SIGNING_KEY = $key
$env:DATABASE_URL = "postgres://devsvault:devsvault_dev_password@localhost:5432/devsvault?sslmode=disable"
```

Ejecuta la API:

```powershell
go run ./apps/api/cmd/api
```

Ejecuta tests criticos:

```powershell
go test ./apps/api/... ./apps/cli/...
```

Ejecuta la CLI:

```powershell
go run ./apps/cli/cmd/devsvault --help
```

Ejecuta el panel web:

```powershell
cd apps/web
npm install
npm run dev
```

## Endpoints actuales

Todos los endpoints viven bajo `/api/v1`.

- `GET /healthz`
- `POST /api/v1/auth/login`
- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces/{id}`
- `PATCH /api/v1/workspaces/{id}`
- `DELETE /api/v1/workspaces/{id}`
- `GET /api/v1/workspaces/{workspaceId}/projects`
- `POST /api/v1/workspaces/{workspaceId}/projects`
- `GET /api/v1/workspaces/{workspaceId}/projects/{id}`
- `PATCH /api/v1/workspaces/{workspaceId}/projects/{id}`
- `DELETE /api/v1/workspaces/{workspaceId}/projects/{id}`
- `GET /api/v1/projects/{id}`
- `GET /api/v1/projects/{projectId}/environments`
- `POST /api/v1/projects/{projectId}/environments`
- `GET /api/v1/projects/{projectId}/environments/{id}`
- `DELETE /api/v1/projects/{projectId}/environments/{id}`
- `GET /api/v1/environments/{id}`
- `GET /api/v1/secrets`
- `POST /api/v1/secrets`
- `GET /api/v1/secrets/resolve?path=workspace/project/env/name`
- `POST /api/v1/secrets/{id}/versions`
- `POST /api/v1/secrets/{id}/versions/{version}/revoke`
- `GET /api/v1/audit/events`

Los listados devuelven solo metadatos. La lectura del valor completo requiere permiso `secret:read_value` y genera auditoria.

## Seguridad del MVP

- Los valores se cifran con envelope encryption: cada version usa una DEK aleatoria y la DEK se cifra con una KEK externa al almacenamiento.
- La base de datos solo debe almacenar ciphertext, nonce, wrapped DEK y metadatos.
- El material criptografico maestro llega por proveedor externo o variable de entorno de desarrollo, nunca por migraciones ni fixtures.
- El logging y la auditoria no incluyen secretos, tokens, passwords, connection strings ni claves.
- Los permisos se evalúan en middleware y en servicios de aplicacion para evitar bypass por interfaz.
- Los secretos se listan como metadatos; el valor solo se resuelve por endpoint explicito y genera auditoria.

## Documentacion

- [docs/architecture.md](docs/architecture.md)
- [docs/database.md](docs/database.md)
- [docs/encryption.md](docs/encryption.md)
- [docs/api.md](docs/api.md)
- [docs/phases.md](docs/phases.md)

## Riesgos y pendientes principales

- Sustituir el issuer de tokens de desarrollo por OIDC real.
- Integrar KMS externo para envolver DEKs.
- Endurecer el almacenamiento local del token de la CLI con keychain del sistema.
- Endurecer rate limiting, mTLS para servicios y proteccion anti-tamper del audit log.
- Ampliar tests de integracion y E2E.