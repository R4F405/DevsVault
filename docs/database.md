# Modelo de datos

El modelo separa tenancy, identidad, permisos, secreto cifrado y trazabilidad.

## Entidades principales

- `workspaces`: raiz de aislamiento organizacional.
- `projects`: agrupacion dentro de un workspace.
- `environments`: `dev`, `staging`, `prod` u otros entornos del proyecto.
- `users`: identidades humanas federables por OIDC.
- `service_identities`: identidades para workloads y automatizaciones.
- `roles`: roles base y roles custom posteriores.
- `policies`: grants por actor, recurso, accion y alcance.
- `secrets`: metadatos y path logico sin valor sensible.
- `secret_versions`: versiones cifradas, estado y DEK envuelta.
- `secret_access_grants`: permisos concretos por secreto cuando haga falta precision extra.
- `audit_logs`: eventos sensibles sanitizados.
- `sessions` y `service_tokens`: autenticacion revocable.

## Reglas de almacenamiento

- `secret_versions.ciphertext` nunca contiene texto plano.
- `secret_versions.wrapped_dek` almacena la DEK cifrada, no una clave maestra.
- KEK, credenciales OIDC, tokens activos y claves maestras no se guardan en migraciones ni seeds.
- Los campos `metadata` no deben contener valores sensibles.

## Estado actual

- `001_init.sql` crea el esquema base, indices y roles iniciales.
- `002_workspaces_projects_environments.sql` agrega campos operativos (`description`, `created_by`, `updated_at`) y ajusta restricciones para proteger recursos con secretos asociados.
- La API usa PostgreSQL cuando existe `DATABASE_URL`; si no existe, arranca con repositorios en memoria para desarrollo puntual y tests.
- Docker Compose monta `db/migrations` en `/docker-entrypoint-initdb.d`, por lo que las migraciones se aplican al crear un volumen nuevo de PostgreSQL.
- Para reaplicar migraciones desde cero en local, hay que destruir el volumen con `docker compose down -v` y levantar de nuevo el stack.

Ver [db/migrations/001_init.sql](../db/migrations/001_init.sql) para el esquema inicial.