---
description: Describe when these instructions should be loaded by the agent based on task context
# applyTo: 'Describe when these instructions should be loaded by the agent based on task context' # when provided, instructions will automatically be added to the request context when the pattern matches an attached file
---

# Instructions para el agente — Secrets Manager

## Objetivo
Construir una aplicación de **gestión de secretos** para desarrollo local y producción, diseñada para sustituir el uso de secretos sensibles en archivos `.env` por un sistema centralizado, cifrado, auditado y con control de acceso granular. [1][2]

## Alcance del producto
- La aplicación **no es un simple CRUD de API keys**; debe comportarse como un secrets manager real. [1]
- Debe soportar **usuarios humanos y servicios/máquinas** con permisos diferenciados. [1]
- Debe separar estrictamente los secretos por **workspace, proyecto y entorno**. [1]
- La primera versión debe centrarse en **secretos estáticos versionados**, dejando secretos dinámicos para una fase posterior. [1][2]

## Principios obligatorios
- Aplicar **least privilege** en todos los accesos a secretos y metadatos. [1]
- Cifrar secretos **en tránsito y en reposo**. [2]
- Evitar por diseño que secretos, tokens o credenciales aparezcan en logs, errores o respuestas innecesarias. [3]
- Tratar el sistema como **infraestructura crítica**, con respaldo, recuperación y tolerancia a fallos. [1]
- Priorizar acceso en runtime frente a copias manuales de secretos entre sistemas. [1]

## Reglas de arquitectura
- Separar el sistema en módulos claros: **Auth**, **Policies/RBAC**, **Secrets**, **Encryption**, **Audit Logs**, **CLI/Agent**, **SDK/API**. [1][4]
- Seguir una arquitectura limpia con capas: `domain`, `application`, `infrastructure`, `interfaces`. [4]
- La base de datos **nunca** debe almacenar secretos en texto plano. [1][2]
- El material criptográfico maestro debe estar **fuera de la base de datos**, idealmente en KMS o un proveedor externo de claves. [2][5]
- Toda lectura, escritura, rotación, revocación o borrado lógico de secretos debe pasar por una capa central de autorización. [1][6]
- Diseñar la app para que la UI, la API y la CLI consuman los mismos servicios de dominio. [4]

## Módulos mínimos

### Auth
- Login para usuarios humanos.
- Identidades de servicio o machine identities.
- Preparado para OIDC/OAuth/SSO. [1][6]

### Policies / RBAC
- Roles base: `admin`, `developer`, `runtime-service`, `auditor`.
- Permisos por workspace, proyecto, entorno y secreto.
- Permisos separados para: ver metadatos, leer valor, escribir, rotar, revocar y administrar accesos. [1]

### Secrets
- Crear secreto.
- Actualizar secreto.
- Versionar secreto.
- Activar una versión concreta.
- Revocar una versión.
- Leer secreto por path lógico.
- Listar secretos sin exponer sus valores. [1][2]

### Encryption
- Usar envelope encryption.
- Cada valor secreto debe cifrarse con una DEK; la DEK debe cifrarse con una KEK o master key externa. [2][5]
- Rotación de claves preparada desde el diseño, aunque no esté automatizada en la primera fase. [2][5]

### Audit Logs
- Registrar eventos de login, logout, acceso a secretos, fallos de autorización, cambios de permisos, rotaciones, revocaciones y cambios administrativos. [3][7]
- Los logs deben excluir cualquier dato sensible. [3]
- Los logs deben ser resistentes a manipulación y con acceso restringido. [3]

### CLI / Agent
- Comandos iniciales: `login`, `pull`, `run`, `inject`.
- Permitir ejecutar apps con secretos efímeros sin persistirlos innecesariamente en disco. [1]
- Cualquier caché local debe ser temporal, cifrada y con expiración corta. [1]

### SDK / API
- API versionada bajo `/api/v1`.
- SDK mínimo para Node y PHP/Laravel.
- Acceso a secretos por path lógico tipo `workspace/project/env/secret-name`. [1][8]

## Modelo de datos mínimo
Tablas o entidades recomendadas:

- `workspaces`
- `projects`
- `environments`
- `users`
- `service_identities`
- `roles`
- `policies`
- `secret_paths` o `secrets`
- `secret_versions`
- `secret_access_grants`
- `audit_logs`
- `sessions` o `service_tokens`

Este modelo debe permitir separar pertenencia, permisos, versionado, acceso runtime y trazabilidad. [1][3]

## Reglas de exposición de secretos
- Nunca devolver el valor completo del secreto en endpoints de listado. [1]
- La UI debe mostrar por defecto solo **metadatos**: nombre, path, entorno, versión activa, fecha de creación, fecha de rotación y último acceso. [3]
- Mostrar el valor completo de un secreto debe requerir permiso explícito y generar un evento de auditoría. [1][3]
- El sistema debe favorecer consumo automático por servicios antes que visualización manual por personas. [1]

## Reglas de logging
- Nunca registrar secretos, access tokens, refresh tokens, passwords, connection strings, private keys ni claves maestras. [3][9]
- Sanitizar mensajes de error y excepciones. [3]
- Incluir contexto útil en logs sin exponer datos sensibles: actor, recurso, acción, resultado, timestamp, IP o contexto técnico cuando aplique. [3][7]
- Proteger logs frente a borrado o alteración no autorizada. [3]

## Reglas de API
- Diseñar endpoints versionados y consistentes. [8]
- Validar inputs de forma estricta en todos los endpoints. [8][6]
- Aplicar rate limiting y middleware de autorización centralizado. [8][6]
- No filtrar detalles internos de cifrado, base de datos o autorización en errores. [3]
- Diferenciar bien entre endpoints de metadatos y endpoints de acceso al valor del secreto. [1]

## Reglas de seguridad runtime
- Las apps cliente deben autenticarse con identidad propia, no con credenciales compartidas entre varios servicios. [1][6]
- El acceso a secretos debe ser temporal cuando sea posible. [1]
- El diseño debe permitir revocar rápidamente accesos o tokens comprometidos. [1][2]
- Preparar el sistema para soportar rotación automática en versiones futuras. [2][5]

## UX mínima del panel
- Panel orientado a seguridad, permisos y trazabilidad, no solo a CRUD. [3]
- Vista clara de quién tiene acceso a qué secreto. [3][6]
- Indicadores de secretos antiguos, sin rotación o con permisos excesivos. [2][5]
- Diferenciar visualmente entre **metadatos**, **políticas** y **valor**. [1]

## Stack recomendado
- **Core API:** Go
- **CLI:** Go
- **Panel web:** Next.js + TypeScript
- **Base de datos:** PostgreSQL
- **Cache/colas opcionales:** Redis
- **Autenticación:** OIDC / Keycloak / Auth0
- **Criptografía:** crypto nativo + envelope encryption + soporte para KMS

Este stack es adecuado para un producto de infraestructura con CLI, runtime seguro y panel web administrativo. [1][2]

## Estructura de carpetas sugerida
```text
/secrets-manager
  /apps
    /api
    /web
    /cli
  /packages
    /sdk-node
    /sdk-php
    /shared-types
  /internal
    /auth
    /policies
    /secrets
    /encryption
    /audit
  /db
    /migrations
    /seeds
  /docs
```

## Reglas de implementación
- No empezar por pantallas o dashboard genérico.
- Empezar por arquitectura, amenazas, modelo de datos, cifrado, permisos y auditoría. [4][1]
- Implementar por fases:
  1. Arquitectura y modelo de datos
  2. Backend base y cifrado
  3. Auth y permisos
  4. Audit log
  5. CLI
  6. Panel web
  7. SDKs
- En cada fase, documentar decisiones, riesgos, límites actuales y siguiente paso. [4]

## Entregables iniciales esperados
1. Diseño de arquitectura.
2. Esquema de base de datos.
3. Endpoints iniciales.
4. Flujo de cifrado.
5. MVP del backend.
6. MVP de la CLI.
7. MVP del panel web.
8. Docker Compose para desarrollo.
9. README con setup local. [4]

## Restricciones importantes
- No usar `.env` para almacenar secretos sensibles del producto final.
- `.env` solo puede contener configuración no crítica del entorno de desarrollo, y aun así debe minimizarse. [1][2]
- No persistir secretos localmente salvo necesidad técnica justificada, y si se hace, debe ser temporal y cifrado. [1]
- No generar código que exponga secretos en tests, fixtures o ejemplos. [3]

## Criterios de éxito
El proyecto se considera bien encaminado si:
- los secretos nunca se almacenan en texto plano,
- el acceso está gobernado por políticas,
- toda acción sensible deja rastro auditable,
- la CLI permite uso local seguro,
- la arquitectura queda preparada para KMS, rotación y escalado futuro. [1][2][5]