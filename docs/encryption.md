# Flujo de cifrado

El MVP usa envelope encryption con AES-256-GCM.

## Escritura

1. Validar workspace, proyecto, entorno, nombre y valor.
2. Autorizar `secret:write` o `secret:rotate`.
3. Generar una DEK aleatoria de 32 bytes por version.
4. Cifrar el valor con AES-256-GCM usando la DEK y AAD derivado de `secret_id`, version y path logico.
5. Obtener KEK activa desde un proveedor externo al almacenamiento.
6. Cifrar la DEK con AES-256-GCM usando la KEK.
7. Guardar `ciphertext`, `nonce`, `wrapped_dek`, `dek_nonce`, `key_id`, version y metadatos.
8. Auditar escritura o rotacion sin registrar el valor.

## Lectura

1. Autorizar `secret:read_value`.
2. Cargar version activa no revocada.
3. Recuperar KEK por `key_id` desde proveedor/KMS.
4. Desencriptar DEK en memoria.
5. Desencriptar valor en memoria y devolverlo solo al endpoint explicito.
6. Borrar referencias al valor lo antes posible y auditar la lectura.

## Preparacion KMS

El servicio de cifrado depende de un puerto `KEKProvider`. El adaptador de desarrollo lee una clave base64 desde el entorno; un adaptador productivo debe implementar el mismo contrato usando AWS KMS, GCP KMS, Azure Key Vault, Vault Transit u otro proveedor.

## Estado actual

- La API exige `DEVSVAULT_MASTER_KEY_B64` y `DEVSVAULT_AUTH_SIGNING_KEY`; ambas deben decodificar a 32 bytes.
- Docker Compose usa claves fijas solo para desarrollo local reproducible.
- En ejecucion manual se debe generar una clave fuera del repo y exportarla en la terminal antes de arrancar la API.
- PostgreSQL almacena solo `ciphertext`, `nonce`, `wrapped_dek`, `dek_nonce`, `key_id` y metadatos de version.

## Riesgos pendientes

- Agregar rotacion de KEK con rewrap de DEKs.
- Agregar key versions y politicas por workspace.
- Usar memoria bloqueada o procesos aislados para reducir exposicion del plaintext en runtime.