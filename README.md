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
  - Eliminación y Reasignación asistida de VPNs inactivas con modal interactivo: permite transferir subredes y dominios a otra VPN con desduplicación estricta (`INSERT OR IGNORE`) y opción de purgar el perfil en NetworkManager (`nmcli connection delete`).
- **🔎 Búsqueda Global Dual (Atajo `Ctrl+K`):**
  - Paleta flotante rápida (`Ctrl+K` o `Cmd+K`) con vista previa instantánea de VPNs, rutas CIDR y dominios.
  - Vista completa (`/search?q=...`) con tablas agrupadas y contadores de coincidencias.
- **⚡ Gestor de Rutas e IPs con Filtros Dinámicos:**
  - Barra ergonómica de filtrado por VPN con contadores en vivo y botón para "Deseleccionar Todo".
  - Buscador instantáneo en caliente.
  - Formulario modal para **editar y actualizar** subredes y descripciones al vuelo.
- **🌐 Gestor de Dominios y Resolución Paralela:**
  - Resolución DNS concurrente en Go nativo mediante goroutines (ultrarrápida).
  - Filtro por VPN con contadores y botón para deseleccionar todo.
  - Inyección dinámica de IPs resueltas a NetworkManager para sitios con restricciones geográficas.
- **💾 Persistencia Dual y Ubicación de Datos (Estándar XDG):**
  - Base de datos SQLite activa ubicada en `~/.config/multivpn/vpn_data.db` (en Linux).
  - Si la base de datos no existe al iniciar, se inicializa y siembra automáticamente desde `data/vpn_data_dump.sql`.
  - Auto-volcado en caliente a `data/vpn_data_dump.sql` para trazabilidad de cambios en Git.
  - Configuración opcional vía variables de entorno en `~/.config/multivpn/config.env`.
- **🚀 Preparado para Multiplataforma:** Motor desacoplado por interfaz `VPNProvider` (Linux hoy, soporte nativo a Windows y BSD en desarrollo).

---

## 📦 Instalación y Descarga

Tienes dos formas de utilizar MultiVPN:

### Opción 1: Instalador rápido con una sola línea (Recomendado en Linux)
Ejecuta en tu terminal para descargar, verificar e instalar la última versión:
```bash
curl -fsSL https://raw.githubusercontent.com/terracenter/multivpn/main/install.sh | bash
```
> El instalador detecta la arquitectura, valida la integridad del ejecutable y lo instala en `~/.local/bin/multivpn` (o en `/usr/local/bin/` si se ejecuta con `sudo`).

---

### Opción 2: Instalación manual paso a paso (Linux)
Si prefieres no usar el script, descarga directamente el ejecutable desde [GitHub Releases](https://github.com/terracenter/multivpn/releases):
```bash
# Descargar con banderas seguras
curl -fsSL -o multivpn https://github.com/terracenter/multivpn/releases/latest/download/multivpn

# Asignar permisos e instalar en tu PATH de usuario
mkdir -p ~/.local/bin
install -m 755 multivpn ~/.local/bin/multivpn
rm -f multivpn

# O en todo el sistema (requiere sudo):
# sudo install -m 755 multivpn /usr/local/bin/multivpn
```

**En Windows:**
Descarga `multivpn.exe` desde [Releases](https://github.com/terracenter/multivpn/releases) y ejecútalo directamente desde la terminal o PowerShell.

---

### Opción 2: Compilar e instalar desde el código fuente (Requiere Go 1.22+)
Si tienes Go instalado y quieres compilarlo localmente:
```bash
git clone https://github.com/terracenter/multivpn.git
cd multivpn

# Compilar el binario local
make build

# Instalar directamente en ~/.local/bin/multivpn
make install
```

> **Nota:** Si tienes `~/.local/bin` en tu `PATH`, podrás ejecutar `multivpn` directamente desde cualquier terminal.

---

## 🚀 Inicio Rápido

Una vez instalado o descargado el binario:

### Ejecutar con WebUI Local (Modo Recomendado)
```bash
multivpn web
# O simplemente:
multivpn
```
El sistema levantará el servidor en `http://127.0.0.1:8765` y abrirá automáticamente tu navegador predeterminado.

### Atajos de Teclado en la WebUI
- **`Ctrl + K`** (o **`Cmd + K`** en macOS): Abre la paleta de búsqueda global instantánea para encontrar VPNs, IPs o dominios.
- **`ESC`**: Cierra el modal activo o la paleta de búsqueda.

### Uso por Línea de Comandos (Headless / SSH)
```bash
# Ver estado de las conexiones
multivpn status

# Conectar una VPN específica
multivpn up FibexValencia

# Conectar todas las VPNs
multivpn up --all

# Desconectar todo
multivpn down

# Escanear y blindar conexiones del sistema (Split Tunneling)
multivpn shield
```

---

## 🛠️ Comandos Makefile Disponibles

| Comando | Descripción |
|---|---|
| `make build` | Compila el CSS Tailwind local y genera el binario `./multivpn` |
| `make install` | Compila y copia el binario a `~/.local/bin/multivpn` |
| `make uninstall` | Elimina el binario de `~/.local/bin/multivpn` |
| `make test` | Ejecuta la suite completa de pruebas unitarias (`go test -v ./...`) |
| `make css` | Recompila `internal/web/static/tailwind.css` usando Tailwind v4 local |
| `make build-windows` | Compilación cruzada para Windows (`multivpn.exe`) |
| `make clean` | Elimina binarios compilados locales |

---

## 📁 Ubicación de Archivos en Linux

- **Base de datos SQLite:** `~/.config/multivpn/vpn_data.db`
- **Archivo de configuración opcional:** `~/.config/multivpn/config.env`
- **Binario instalado por usuario:** `~/.local/bin/multivpn`
- **Dump de sincronización Git:** `<repositorio>/data/vpn_data_dump.sql`

---

## 🤝 Contribución y Comunidad

Este proyecto es **100% Software Libre** bajo la licencia **GNU General Public License v3.0 (GPLv3)**.
Todas las ideas de mejora, sugerencias, adaptaciones para Windows/BSD y reportes de errores son más que bienvenidas. 

Por favor, revisa [`CONTRIBUTING.md`](CONTRIBUTING.md) para saber cómo colaborar.
