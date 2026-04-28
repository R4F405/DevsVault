# DevsVault Secrets Manager

MVP de un secrets manager para desarrollo local y produccion. El sistema esta pensado para almacenar secretos estaticos versionados, recuperarlos por path logico, aplicar permisos por recurso y auditar acciones sensibles sin depender de archivos `.env` para valores sensibles.

## Estado del MVP

- Fase 1: arquitectura, modelo de datos, flujo de cifrado y permisos documentados.
- Fase 2: backend Go ejecutable con cifrado envelope AES-256-GCM, repositorios en memoria y tests criticos.
- Fase 3: auth preparada para OIDC con tokens firmados de desarrollo y autorizacion centralizada.
- Fase 4: audit log sanitizado para lecturas, escrituras, revocaciones, logins y denegaciones.
- Fase 5: CLI Go con `login`, `pull`, `run` e `inject`.
- Fase 6: panel Next.js + TypeScript orientado a secretos, accesos y auditoria.
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

El backend usa capas por modulo: `domain`, `application`, `infrastructure` e `interfaces`. Los modulos iniciales son `auth`, `policies`, `secrets`, `encryption` y `audit`.

## Setup local

Requisitos:

- Go 1.22+
- Node.js 20+
- Docker Desktop

Levanta PostgreSQL:

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
```

Ejecuta la API:

```powershell
go run ./apps/api/cmd/api
```

Ejecuta tests criticos:

```powershell
go test ./apps/api/...
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

## Endpoints iniciales

Todos los endpoints viven bajo `/api/v1`.

- `GET /healthz`
- `POST /api/v1/auth/login`
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

## Documentacion

- [docs/architecture.md](docs/architecture.md)
- [docs/database.md](docs/database.md)
- [docs/encryption.md](docs/encryption.md)
- [docs/api.md](docs/api.md)
- [docs/phases.md](docs/phases.md)

## Riesgos y pendientes principales

- Conectar los repositorios en memoria a PostgreSQL usando las migraciones existentes.
- Sustituir el issuer de tokens de desarrollo por OIDC real.
- Integrar KMS externo para envolver DEKs.
- Endurecer rate limiting, mTLS para servicios y proteccion anti-tamper del audit log.
- Ampliar tests de integracion y E2E.