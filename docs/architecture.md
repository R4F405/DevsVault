# Arquitectura

DevsVault se organiza como una plataforma de infraestructura critica. El dominio no depende de HTTP, SQL, Next.js ni CLI. La API, el panel, la CLI y los SDKs deben consumir los mismos casos de uso.

## Modulos

- `auth`: usuarios humanos, identidades de servicio, sesiones y preparacion OIDC.
- `policies`: roles, grants y decisiones de autorizacion por recurso.
- `secrets`: paths logicos, versiones, revocacion y lectura controlada.
- `encryption`: envelope encryption, proveedores KEK y preparacion KMS.
- `audit`: eventos inmutables y sanitizados de acciones sensibles.

## Capas

Cada modulo sigue esta convencion:

- `domain`: entidades, tipos y errores de negocio.
- `application`: casos de uso, puertos y orquestacion.
- `infrastructure`: adaptadores concretos como memoria, SQL, KMS u OIDC.
- `interfaces`: HTTP, CLI handlers o DTOs externos.

## Flujo de una lectura

1. El cliente llama `GET /api/v1/secrets/resolve?path=workspace/project/env/name` con un token.
2. Middleware autentica y crea un `Actor` sin exponer el token.
3. El servicio de secretos valida el path logico.
4. `policies` decide si el actor tiene `secret:read_value` sobre workspace, proyecto, entorno y secreto.
5. `secrets` carga solo la version activa no revocada.
6. `encryption` desenvuelve la DEK mediante KEK/KMS y descifra el valor en memoria.
7. `audit` registra lectura exitosa o denegada sin incluir el valor.
8. La respuesta contiene el valor solo en este endpoint de acceso explicito.

## Permisos base

- `admin`: administra workspace, politicas, secretos y auditoria.
- `developer`: lista metadatos y escribe secretos en entornos permitidos.
- `runtime-service`: lee valores de secretos concretos para ejecucion.
- `auditor`: consulta auditoria y metadatos sin leer valores.

Acciones granulares iniciales:

- `secret:list_metadata`
- `secret:read_value`
- `secret:write`
- `secret:rotate`
- `secret:revoke`
- `access:manage`
- `audit:read`

## Decisiones

- El MVP implementa repositorios en memoria para arrancar y testear servicios sin infraestructura externa; las migraciones SQL definen el contrato persistente.
- El cifrado se implementa antes que la UI para evitar un CRUD inseguro.
- La autorizacion vive en un servicio central y tambien se aplica en HTTP para mantener least privilege.
- Las respuestas de listado nunca contienen valores completos de secretos.

## Riesgos

- La memoria no es persistente y debe reemplazarse por PostgreSQL antes de un despliegue real.
- Las variables de entorno de desarrollo para KEK y firma no equivalen a KMS ni OIDC productivo.
- Falta proteccion criptografica anti-manipulacion para audit logs, como hash chaining o almacenamiento WORM.