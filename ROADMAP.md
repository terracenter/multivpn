# Roadmap — multivpn

## Estado actual

- Estado: activo
- Último hito: versión 1.0.0 publicada (WebUI local, blindaje de gateway, búsqueda global, resolución DNS concurrente)
- Próxima acción: definir el alcance del soporte de Windows y BSD, hoy marcado como en desarrollo (WIP)

## Fase 0 — Base operativa

- [x] Ficha del proyecto en el vault.
- [x] Documentación mínima del repo.
- [x] Validación local mínima documentada (`make test`).
- [x] `.gitignore` protege secretos y artefactos generados.

## Fase 1 — Funcionalidad inicial

- [x] WebUI local con SQLite, rutas y dominios por VPN.
- [x] Blindaje de Default Gateway (split-tunneling estricto) sobre NetworkManager.
- [ ] Soporte de Windows (WIP).
- [ ] Soporte de BSD (WIP).

## Bloqueos

- Sin bloqueos conocidos.
