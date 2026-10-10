# Guía de Contribución a MultiVPN 🌐

¡Gracias por tu interés en mejorar MultiVPN! Este es un proyecto de **Software Libre (GPLv3)**, totalmente gratuito para la comunidad. Si tienes ideas de mejora, nuevas funciones (como drivers de red para Windows o BSD) o correcciones de bugs, ¡serán muy bien recibidas!

---

## 💡 ¿Cómo puedes colaborar?

1. **Reportar Errores (Bugs):**
   - Abre un *Issue* describiendo tu sistema operativo, versión de NetworkManager/VPN, qué esperabas y qué sucedió.
2. **Sugerir Mejoras o Funcionalidades:**
   - Si tienes ideas para hacer la interfaz más amigable o añadir características útiles, abre una discusión o *Issue* con la etiqueta `enhancement`.
3. **Enviar Pull Requests (PR):**
   - Toda corrección o aporte de código es bienvenido. Mantén los PRs pequeños y enfocados en un solo propósito.

---

## 🛠️ Estándares del Proyecto

- **Idioma de documentación y commits:** Español.
- **Identificadores de código (Go, variables, funciones):** Inglés americano idiomático.
- **Formato de Commits (Conventional Commits):**
  ```text
  feat(modulo): descripción imperativa
  fix(modulo): descripción imperativa
  docs(modulo): descripción imperativa
  ```
- **Pruebas y calidad:** Todo cambio relevante debe compilar limpiamente (`go build ./...`) y pasar los tests correspondientes (`go test ./...`).
