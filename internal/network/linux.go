package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/terracenter/multivpn/internal/models"
)

// LinuxManager implementa VPNProvider usando NetworkManager (nmcli) en Linux
type LinuxManager struct{}

// NewLinuxManager crea una nueva instancia del gestor Linux
func NewLinuxManager() *LinuxManager {
	return &LinuxManager{}
}

// ScanConnections detecta perfiles de VPN en NetworkManager (vpn y wireguard)
func (m *LinuxManager) ScanConnections() ([]models.Connection, error) {
	cmd := exec.Command("nmcli", "-g", "NAME,TYPE", "connection", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando nmcli: %w", err)
	}

	activeMap, _ := m.getActiveConnectionsMap()

	var connections []models.Connection
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			name := parts[0]
			connType := parts[1]

			if connType == "vpn" || connType == "wireguard" {
				hasRisk, _ := m.HasDefaultGatewayRisk(name)
				isActive := activeMap[name]

				connections = append(connections, models.Connection{
					Name:              name,
					Type:              connType,
					IsActive:          isActive,
					HasDefaultGateway: hasRisk,
				})
			}
		}
	}

	return connections, nil
}

// ShieldConnection aplica blindaje de Default Gateway (Split Tunneling obligatorio)
func (m *LinuxManager) ShieldConnection(connName string) error {
	cmd := exec.Command("nmcli", "connection", "modify", connName,
		"ipv4.never-default", "yes",
		"ipv4.ignore-auto-routes", "yes",
		"ipv6.method", "disabled",
		"ipv6.routes", "",
		"ipv6.addresses", "",
		"ipv6.gateway", "")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("falla blindando '%s': %s (%w)", connName, strings.TrimSpace(string(output)), err)
	}
	return nil
}

// HasDefaultGatewayRisk comprueba si una VPN tiene never-default en 'no' (riesgo de fuga de gateway)
func (m *LinuxManager) HasDefaultGatewayRisk(connName string) (bool, error) {
	cmd := exec.Command("nmcli", "-g", "ipv4.never-default", "connection", "show", connName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	val := strings.TrimSpace(string(output))
	// Si es "no" o está vacío, significa que el sistema usará el gateway por defecto de la VPN
	return val == "no" || val == "", nil
}

// IsConnectionActive verifica si una conexión está arriba
func (m *LinuxManager) IsConnectionActive(connName string) (bool, error) {
	activeMap, err := m.getActiveConnectionsMap()
	if err != nil {
		return false, err
	}
	return activeMap[connName], nil
}

// Up aplica rutas estáticas + dominios resueltos y levanta la VPN
func (m *LinuxManager) Up(conn models.Connection, routes []string, resolvedIPs []string) error {
	var routeElements []string

	// Agregar rutas CIDR estáticas
	for _, r := range routes {
		r = strings.TrimSpace(r)
		if r != "" {
			routeElements = append(routeElements, fmt.Sprintf("%s 0.0.0.0", r))
		}
	}

	// Agregar IPs de dominios resueltos como /32
	for _, ip := range resolvedIPs {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			routeElements = append(routeElements, fmt.Sprintf("%s/32 0.0.0.0", ip))
		}
	}

	routeStr := strings.Join(routeElements, ", ")

	// Parámetros de NetworkManager para máxima estabilidad y split-tunneling
	args := []string{"connection", "modify", conn.Name,
		"ipv4.routes", routeStr,
		"ipv4.never-default", "yes",
		"ipv4.ignore-auto-routes", "yes",
		"ipv6.method", "disabled",
		"ipv6.routes", "",
		"ipv6.addresses", "",
		"ipv6.gateway", "",
	}

	if conn.IsSalida {
		// En perfiles de salida no sobreescribir el DNS del sistema para no alterar AdGuard o DNS local
		args = append(args, "ipv4.ignore-auto-dns", "yes")
	}

	modCmd := exec.Command("nmcli", args...)
	if out, err := modCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error modificando rutas en NetworkManager: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	// Levantar conexión
	upCmd := exec.Command("nmcli", "connection", "up", conn.Name)
	if out, err := upCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error levantando conexión '%s': %s (%w)", conn.Name, strings.TrimSpace(string(out)), err)
	}

	return nil
}

// Down desconecta el perfil
func (m *LinuxManager) Down(connName string) error {
	cmd := exec.Command("nmcli", "connection", "down", connName)
	_ = cmd.Run()
	return nil
}

// DeleteConnectionProfile elimina el perfil en NetworkManager
func (m *LinuxManager) DeleteConnectionProfile(connName string) error {
	cmd := exec.Command("nmcli", "connection", "delete", connName)
	return cmd.Run()
}

// getActiveConnectionsMap obtiene un mapa de nombres de conexiones activas
func (m *LinuxManager) getActiveConnectionsMap() (map[string]bool, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,STATE", "connection", "show", "--active")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	activeMap := make(map[string]bool)
	lines := strings.Split(out.String(), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Split(l, ":")
		if len(parts) >= 1 && parts[0] != "" {
			activeMap[parts[0]] = true
		}
	}
	return activeMap, nil
}
