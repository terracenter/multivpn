# Estándar de presentación GitHub

Este estándar define cómo deben verse los repos del workspace en GitHub. Se basa en la comparación de `terracenter/security-manager`, `davila7/claude-code-templates`, `agent-orchestrator` y `sge-observer-llm`.

## Objetivo

Que cada repo sea entendible y profesional en menos de 60 segundos:

1. Qué es.
2. Para quién es.
3. Estado real.
4. Cómo probarlo.
5. Cómo se opera.
6. Qué está pendiente.
7. Cómo contribuir o auditar cambios.

## README mínimo profesional

Todo `README.md` debe incluir:

- selector de idioma arriba si existe traducción;
- badges de licencia, estado, stack y política de PRs/checks;
- título claro;
- descripción corta tipo pitch;
- tabla de estado actual;
- sección “Qué resuelve”;
- quickstart verificable;
- arquitectura o estructura resumida;
- enlaces a ROADMAP, RELEASES, SECURITY y docs operativas;
- política de documentación como changelog operativo;
- licencia.

## ROADMAP

Debe indicar:

- visión;
- principios no negociables;
- estado actual;
- fases con checkboxes;
- bloqueos activos;
- política de actualización.

No se marca una fase como completa si no existe PR/issue mergeado/cerrado que lo respalde.

## RELEASES / CHANGELOG

Debe registrar cambios operativos relevantes, incluso si todavía no hay release formal:

- comandos nuevos;
- endpoints nuevos;
- cambios de seguridad;
- cambios de despliegue;
- decisiones de arquitectura;
- documentación obligatoria actualizada.

## SECURITY

Debe explicar:

- cómo reportar vulnerabilidades;
- qué no debe subirse a git;
- estado de soporte;
- si el repo toca producción, credenciales, infraestructura o agentes remotos.

## CONTRIBUTING

Debe exigir:

- PRs, no commits directos a `main`;
- tests/checks cuando aplique;
- documentación actualizada o `Docs: no aplica` justificado;
- respeto a licencia y atribución.

## Plantillas GitHub

Todo repo mantenido debe tener:

- `.github/pull_request_template.md`;
- issue template de bug;
- issue template de feature;
- issue template de documentación si el proyecto es público o de uso operativo.

## Regla permanente

Un issue, PR o entregable no se considera cerrado hasta que la documentación correspondiente esté actualizada y verificable en GitHub, o hasta que el PR justifique explícitamente por qué no aplica.
