# 🌐 MultiVPN

![Licencia](https://img.shields.io/badge/licencia-GPLv3-green)
![Estado](https://img.shields.io/badge/estado-En_Desarrollo-orange)
![Stack](https://img.shields.io/badge/stack-Go_%7C_HTMX_%7C_Alpine.js_%7C_Tailwind-informational)
![Plataformas](https://img.shields.io/badge/plataformas-Linux_%7C_Windows_(WIP)_%7C_BSD_(WIP)-blue)

Gestor inteligente de múltiples conexiones VPN con **WebUI local embebida**, persistencia dual en **SQLite** y **blindaje automático de Default Gateway** (Split Tunneling estricto).

Desarrollado en **Go** como un binario estático único, sin dependencias externas pesadas ni cadenas de Node.js.

---

## 🌟 Características Principales

- **🛡️ Blindaje Automático de Default Gateway:** Al detectar cualquier conexión creada en `NetworkManager`, `nmtui` o perfiles del sistema, desactiva automáticamente el gateway remoto (`ipv4.never-default yes`). Tu tráfico de internet habitual **nunca se interrumpe ni se desvía**.
- **🌐 WebUI Local Amigable (HTMX + Alpine.js):** Ejecuta `./multivpn web` y se abrirá automáticamente tu navegador con un panel visual interactivo:
  - Tarjetas de conexión con estado en tiempo real (🟢 Activa / ⚪ Inactiva).
  - Selector individual y en lote ("Conectar Todas" / "Desconectar Todas").
- **⚡ Gestor de Rutas e IPs con Edición en Vivo:**
  - Tabla navegable con buscador instantáneo en caliente.
  - Formulario modal para **editar y actualizar** subredes y descripciones al vuelo (sin tener que borrar y recrear).
- **🔎 Gestor de Dominios y Resolución Paralela:**
  - Resolución DNS concurrente en Go nativo mediante goroutines (ultrarrápida).
  - Inyección dinámica de IPs resueltas a NetworkManager para sitios con restricciones geográficas.
- **💾 Persistencia Dual (SQLite + Git):**
  - Almacenamiento activo en base de datos SQLite sin CGO (`modernc.org/sqlite`).
  - Auto-volcado en caliente a `data/vpn_data_dump.sql` para trazabilidad de cambios en Git.
- **🚀 Preparado para Multiplataforma:** Motor desacoplado por interfaz `VPNProvider` (Linux hoy, soporte nativo a Windows y BSD en desarrollo).

---

## 🚀 Inicio Rápido

### Ejecutar con WebUI Local (Modo Recomendado)
```bash
./multivpn web
# O simplemente:
./multivpn
```
El sistema levantará el servidor en `http://127.0.0.1:8765` y abrirá automáticamente tu navegador predeterminado.

### Uso por Línea de Comandos (Headless / SSH)
```bash
# Ver estado de las conexiones
./multivpn status

# Conectar una VPN específica
./multivpn up FibexValencia

# Conectar todas las VPNs
./multivpn up --all

# Desconectar todo
./multivpn down

# Escanear y blindar conexiones del sistema (Split Tunneling)
./multivpn shield
```

---

## 🛠️ Compilación y Desarrollo

Requisitos:
- **Go 1.22+**
- **NetworkManager (`nmcli`)** (en Linux)

```bash
# Compilar binario nativo
make build

# Ejecutar tests
make test

# Compilación cruzada para Windows (.exe)
make build-windows
```

---

## 🤝 Contribución y Comunidad

Este proyecto es **100% Software Libre** bajo la licencia **GNU General Public License v3.0 (GPLv3)**.
Todas las ideas de mejora, sugerencias, adaptaciones para Windows/BSD y reportes de errores son más que bienvenidas. 

Por favor, revisa [`CONTRIBUTING.md`](CONTRIBUTING.md) para saber cómo colaborar.
