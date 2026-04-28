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

Ver [db/migrations/001_init.sql](../db/migrations/001_init.sql) para el esquema inicial.