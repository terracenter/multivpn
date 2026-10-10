# Política de aprobación de PRs y autoridad de entrega

> [!DANGER]
> Un agente, bot o LLM no debe aprobar su propio trabajo ni saltarse reglas de protección de rama. La seguridad y la separación de responsabilidades tienen prioridad sobre velocidad.

## Problema

GitHub no permite que el autor de un Pull Request apruebe su propio PR cuando la rama exige aprobación. Eso es correcto: evita que la misma identidad sea autora y autoridad de entrega.

## Regla del workspace

| Situación | Regla |
|---|---|
| PR creado por humano | Debe aprobarlo otro humano/bot autorizado distinto, si la protección exige review. |
| PR creado por agente usando token del humano | GitHub lo considera autor humano; ese mismo humano/token no puede autoaprobar. |
| PR creado por bot oficial | Freddy u otro maintainer con permisos revisa y aprueba. |
| Repo single-maintainer sin reviewer real | No usar approvals falsos; usar CI fuerte, 4R, evidencia y reglas de protección compatibles. |

## Prohibido

- Desactivar protección de rama solo para mergear rápido.
- Usar una cuenta secundaria como “sello” sin revisión real.
- Forzar push a `main` para evitar review.
- Marcar una revisión automática como aprobación humana si no lo es.
- Usar resultados de LLM como permiso de entrega.

## Permitido

- Usar modelos para revisión informativa 4R.
- Usar `orq review 4r` para generar checklist.
- Usar CI, SAST, pruebas y evidencia como señales de confianza.
- Crear una GitHub App o machine user oficial para autoría técnica, siempre con revisión humana separada.
- Ajustar protección de rama de forma permanente y documentada cuando el repo sea single-maintainer.

## Diseño recomendado

### Repos con separación de roles

1. El agente/bot crea rama y PR.
2. CI ejecuta quality gates.
3. `orq review 4r` genera evidencia.
4. Freddy u otro maintainer revisa y aprueba.
5. Merge solo si GitHub reporta checks verdes y review válida.

### Repos personales/single-maintainer

1. El agente crea PR o rama de trabajo.
2. `Require a pull request before merging` permanece activo.
3. `Required approvals` debe ser `0` si no existe segundo maintainer real.
4. CI fuerte obligatorio.
5. Revisión 4R obligatoria como evidencia.
6. Merge permitido sin approval externo solo si la protección del repo fue diseñada así de forma explícita y permanente.
7. Cambios de seguridad/deploy/secrets siguen requiriendo pausa y confirmación humana.

Configuración mínima recomendada:

| Regla | Valor |
|---|---|
| Require a pull request before merging | Activado |
| Required approvals | `0` en modo single-maintainer |
| Require status checks to pass | Activado |
| Require conversation resolution | Activado |
| Block force pushes | Activado |
| Block deletions | Activado |

## Principio tomado de harnesses existentes

El harness puede revisar y producir evidencia, pero la entrega sigue la política ordinaria del repositorio. Ningún agente debe inventar una ruta de aprobación alternativa.

## Criterios de aceptación

- Dado un PR creado por un agente, cuando GitHub exige review, entonces no se intenta autoaprobar.
- Dado un repo con protección, cuando un PR no cumple los requisitos, entonces queda bloqueado y documentado.
- Dado un repo single-maintainer, cuando no exista reviewer real, entonces la política debe definirse explícitamente sin approvals falsos.
- Dado un cambio sensible, cuando haya duda de seguridad, entonces se consulta otro modelo/humano antes de entregar.
