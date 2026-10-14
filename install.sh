#!/usr/bin/env bash
# ==============================================================================
# MultiVPN — Script de Instalación Automatizado
# Repositorio: https://github.com/terracenter/multivpn
# Licencia: GPLv3
# ==============================================================================

set -euo pipefail

# Colores y formatos
if [[ -t 1 ]]; then
    BOLD="\033[1m"
    GREEN="\033[32m"
    CYAN="\033[36m"
    YELLOW="\033[33m"
    RED="\033[31m"
    RESET="\033[0m"
else
    BOLD=""
    GREEN=""
    CYAN=""
    YELLOW=""
    RED=""
    RESET=""
fi

info() {
    echo -e "${CYAN}${BOLD}[INFO]${RESET} $*"
}

success() {
    echo -e "${GREEN}${BOLD}[OK]${RESET} $*"
}

warn() {
    echo -e "${YELLOW}${BOLD}[AVISO]${RESET} $*"
}

error() {
    echo -e "${RED}${BOLD}[ERROR]${RESET} $*" >&2
}

die() {
    error "$*"
    exit 1
}

# 1. Detección de Arquitectura y SO
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [[ "$OS" != "linux" ]]; then
    die "Este instalador es exclusivo para Linux. Para Windows, descarga 'multivpn.exe' directamente desde GitHub Releases."
fi

case "$ARCH" in
    x86_64|amd64)
        BINARY_ARCH="amd64"
        ;;
    *)
        die "Arquitectura no soportada actualmente para binarios precompilados: $ARCH. Considera compilar desde código fuente con Go."
        ;;
esac

# 2. Definición del directorio de instalación
if [[ -n "${MULTIVPN_INSTALL_DIR:-}" ]]; then
    INSTALL_DIR="$MULTIVPN_INSTALL_DIR"
elif [[ "$(id -u)" -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
fi

REPO="terracenter/multivpn"
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/multivpn"

echo -e "${BOLD}🌐 Instalador Oficial de MultiVPN${RESET}"
echo -e "   Repositorio: ${CYAN}https://github.com/${REPO}${RESET}"
echo -e "   Destino:     ${CYAN}${INSTALL_DIR}/multivpn${RESET}\n"

# 3. Directorio temporal de trabajo
TMP_DIR="$(mktemp -d -t multivpn-install-XXXXXX)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

info "Descargando última versión de MultiVPN..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/multivpn"; then
    die "Fallo al descargar desde: $DOWNLOAD_URL\nVerifica tu conexión o los límites de la API de GitHub."
fi

# 4. Validar integridad básica del binario
if [[ ! -s "$TMP_DIR/multivpn" ]]; then
    die "El archivo descargado está vacío."
fi

if ! file "$TMP_DIR/multivpn" | grep -q "ELF 64-bit"; then
    die "El archivo descargado no parece ser un ejecutable ELF de 64 bits válido."
fi

chmod +x "$TMP_DIR/multivpn"

info "Verificando funcionamiento del binario descargado..."
VERSION_STR="$("$TMP_DIR/multivpn" --version 2>/dev/null || echo "MultiVPN (versión desconocida)")"
success "Binario funcional: $VERSION_STR"

# 5. Resolver conflictos de destino (evitar error de directorio previo)
if [[ -d "$INSTALL_DIR/multivpn" ]]; then
    warn "Se detectó que '$INSTALL_DIR/multivpn' existe como directorio."
    warn "Eliminando carpeta conflictiva para permitir la instalación del binario..."
    rm -rf "$INSTALL_DIR/multivpn"
fi

# Asegurar que el directorio contenedor exista
mkdir -p "$INSTALL_DIR"

# 6. Instalar el binario
info "Instalando en ${INSTALL_DIR}/multivpn..."
if command -v install >/dev/null 2>&1; then
    install -m 755 "$TMP_DIR/multivpn" "$INSTALL_DIR/multivpn"
else
    cp "$TMP_DIR/multivpn" "$INSTALL_DIR/multivpn"
    chmod 755 "$INSTALL_DIR/multivpn"
fi

success "¡MultiVPN instalado con éxito en ${INSTALL_DIR}/multivpn!"

# 7. Verificar si el directorio está en el PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*)
        # En el PATH
        ;;
    *)
        echo ""
        warn "La ruta ${INSTALL_DIR} no está actualmente en tu variable PATH."
        echo "   Para poder ejecutar 'multivpn' desde cualquier terminal, añade la siguiente línea a tu ~/.bashrc o ~/.zshrc:"
        echo -e "   ${CYAN}export PATH=\"${INSTALL_DIR}:\$PATH\"${RESET}"
        echo ""
        ;;
esac

echo ""
echo -e "${GREEN}${BOLD}🚀 ¿Cómo empezar?${RESET}"
echo -e "   Inicia la WebUI local ejecutando:  ${CYAN}multivpn web${RESET}"
echo -e "   Consulta el estado de conexiones:  ${CYAN}multivpn status${RESET}"
echo -e "   Blindar split-tunneling:           ${CYAN}multivpn shield${RESET}"
echo ""
