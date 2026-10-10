# Política de seguridad

## Reporte de vulnerabilidades

Reportar vulnerabilidades directamente a Freddy/maintainer del proyecto. No abrir issues públicos con secretos, credenciales, tokens, DSN reales o pasos explotables contra producción.

## Manejo de secretos

- No commitear `.env`, tokens, claves, dumps, logs con credenciales ni bases locales.
- Usar ejemplos redactados: `<token>`, `<dsn>`, `<password>`.
- Los archivos de configuración local deben estar fuera del repo o ignorados por `.gitignore`.

## Validación mínima

Todo cambio sensible debe indicar:

- qué se cambió;
- qué comando lo verificó;
- qué riesgo queda pendiente.
