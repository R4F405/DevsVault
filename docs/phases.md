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
- Repositorios en memoria para MVP ejecutable y tests rapidos.

Riesgos:

- Falta adaptador PostgreSQL y rotacion de KEK.

Pendientes:

- Implementar repos SQL transaccionales.

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

- `pull`, `run` e `inject` consumen API por path logico.
- No se escriben secretos a disco salvo salida explicita del usuario.

Riesgos:

- El almacenamiento local del token requiere hardening por SO.

Pendientes:

- Cache cifrada con expiracion corta y keychain del sistema.

## Fase 6: panel web

Decisiones:

- Primera pantalla centrada en secretos, permisos y auditoria.
- No se muestran valores completos en listados.

Riesgos:

- Falta integracion real con API y flujos de OIDC.

Pendientes:

- Formularios conectados, estados de error y pruebas visuales.

## Fase 7: SDKs

Decisiones:

- SDKs minimos para resolver secretos por path.
- No loguear tokens ni valores.

Riesgos:

- Falta retry, backoff, cache temporal y tracing seguro.

Pendientes:

- Laravel provider completo, Node ESM/CJS empaquetado y tests.