# Fases

## Fase 1: arquitectura y modelo de datos

Decisiones:

- Arquitectura limpia por modulos y capas.
- Path logico `workspace/project/environment/name` como contrato externo.
- Modelo SQL preparado para versionado, RBAC, identidades de servicio y auditoria.

Riesgos:

- El modelo puede necesitar particionado de auditoria y politicas mas expresivas.

Pendientes:

- Agregar diagramas y threat model formal.

## Fase 2: backend base y cifrado

Decisiones:

- AES-256-GCM para secreto y DEK envuelta.
- `KEKProvider` como puerto para KMS externo.
- Repositorios PostgreSQL para ejecucion local completa y repositorios en memoria para fallback y tests rapidos.

Riesgos:

- Falta rotacion de KEK y hardening operativo de Postgres.

Pendientes:

- Rewrap de DEKs, backups documentados y tests de integracion contra Postgres.

## Fase 3: auth y permisos

Decisiones:

- Token issuer de desarrollo desacoplado de OIDC futuro.
- Autorizacion por acciones granulares y alcance de recurso.

Riesgos:

- El login de desarrollo no debe usarse en produccion.

Pendientes:

- OIDC discovery, JWKS validation, refresh flows y service tokens hashed.

## Fase 4: audit log

Decisiones:

- Eventos estructurados sin secretos.
- Auditoria en exitos y denegaciones.

Riesgos:

- Los logs aun no tienen hash chain ni almacenamiento WORM.

Pendientes:

- Firma de eventos, retencion y exportacion SIEM.

## Fase 5: CLI

Decisiones:

- `login` guarda una sesion local y `logout` la elimina.
- `workspaces`, `projects` y `environments` gestionan recursos via API.
- `secrets list|get|set|rotate|revoke` consume la API por path logico o ID.
- `run -- <command>` inyecta secretos como variables de entorno sin escribir valores a disco.

Riesgos:

- El almacenamiento local del token requiere hardening por SO.

Pendientes:

- Cache cifrada con expiracion corta y keychain del sistema.

## Fase 6: panel web

Decisiones:

- Panel Next.js conectado a la API para login, dashboard, recursos, secretos y auditoria.
- No se muestran valores completos en listados.

Riesgos:

- Falta OIDC real, manejo fino de permisos por vista y pruebas visuales/E2E.

Pendientes:

- Pulir estados vacios/error, permisos en UI y pruebas visuales.

## Fase 7: SDKs

Decisiones:

- SDKs minimos para resolver secretos por path.
- No loguear tokens ni valores.

Riesgos:

- Falta retry, backoff, cache temporal y tracing seguro.

Pendientes:

- Laravel provider completo, Node ESM/CJS empaquetado y tests.