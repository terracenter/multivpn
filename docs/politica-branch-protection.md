# Política de protección de ramas

## Objetivo

Mantener entregas seguras en repos con agentes sin inventar aprobaciones falsas.

## Modo single-maintainer seguro

Cuando el repositorio tenga un solo maintainer real, la protección de `main` debe configurarse así:

| Regla | Valor |
|---|---|
| Require a pull request before merging | Activado |
| Required approvals | `0` |
| Require status checks to pass | Activado |
| Require branches to be up to date before merging | Activado si el flujo lo permite |
| Require conversation resolution | Activado |
| Block force pushes | Activado |
| Block deletions | Activado |

## Por qué approvals en 0

Si no existe segundo maintainer real, exigir una aprobación genera una puerta falsa: bloquea el trabajo legítimo o empuja a usar cuentas sello. Eso no mejora la seguridad.

La seguridad debe venir de evidencia reproducible:

- CI verde;
- pruebas locales;
- `orq repo check`;
- `orq safety check`;
- `orq review 4r`;
- recibos RDD cuando estén disponibles.

## Prohibido

- Desactivar protecciones temporalmente para mergear rápido.
- Usar cuentas secundarias como sello sin revisión real.
- Autoaprobar PRs creados por la misma identidad.
- Tratar una revisión de LLM como aprobación humana.

## Cambios de alto riesgo

Pausar y confirmar explícitamente antes de mergear si el PR toca:

- auth/autorización;
- secretos/tokens;
- firewall/red pública;
- deploy/infraestructura;
- migraciones SQL;
- dependencias críticas;
- borrados masivos o cambios irreversibles.
