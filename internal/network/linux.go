package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/terracenter/multivpn/internal/models"
)

// LinuxManager implementa VPNProvider usando NetworkManager (nmcli) en Linux
type LinuxManager struct {
	mu         sync.Mutex
	blackholes map[string][]string // connName -> []ipv6_blackholed
}

// NewLinuxManager crea una nueva instancia del gestor Linux
func NewLinuxManager() *LinuxManager {
	return &LinuxManager{
		blackholes: make(map[string][]string),
	}
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

// hasIPv6Support verifica si el perfil de conexión tiene soporte o dirección IPv6
func (m *LinuxManager) hasIPv6Support(connName string) bool {
	cmd := exec.Command("nmcli", "-g", "ipv6.method,ipv6.addresses", "connection", "show", connName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	val := strings.TrimSpace(string(output))
	if strings.Contains(val, "disabled") || val == "" || val == ":" {
		return false
	}
	return true
}

// applyBlackholeIPv6 aplica rutas 'unreachable' a IPs IPv6 para forzar a las aplicaciones a usar IPv4 (Blindaje Anti-Fugas)
func (m *LinuxManager) applyBlackholeIPv6(connName string, ips []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Limpiar registros previos si los hubiera
	if old, exists := m.blackholes[connName]; exists {
		for _, ip := range old {
			_ = exec.Command("ip", "-6", "route", "del", ip+"/128", "type", "unreachable").Run()
			_ = exec.Command("sudo", "ip", "-6", "route", "del", ip+"/128", "type", "unreachable").Run()
		}
	}

	var applied []string
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		cmd := exec.Command("ip", "-6", "route", "replace", ip+"/128", "type", "unreachable")
		if err := cmd.Run(); err != nil {
			_ = exec.Command("sudo", "ip", "-6", "route", "replace", ip+"/128", "type", "unreachable").Run()
		}
		applied = append(applied, ip)
	}
	m.blackholes[connName] = applied
}

// clearBlackholeIPv6 elimina las rutas unreachable registradas para la conexión
func (m *LinuxManager) clearBlackholeIPv6(connName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ips, exists := m.blackholes[connName]; exists {
		for _, ip := range ips {
			_ = exec.Command("ip", "-6", "route", "del", ip+"/128", "type", "unreachable").Run()
			_ = exec.Command("sudo", "ip", "-6", "route", "del", ip+"/128", "type", "unreachable").Run()
		}
		delete(m.blackholes, connName)
	}
}

// Up aplica rutas estáticas + dominios resueltos y levanta la VPN adaptándose inteligentemente a IPv4 o Dual-Stack
func (m *LinuxManager) Up(conn models.Connection, routes []string, targets models.ResolvedTargets) error {
	var ipv4Routes []string
	var ipv6Routes []string

	// Separar rutas CIDR estáticas en IPv4 e IPv6
	for _, r := range routes {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if strings.Contains(r, ":") {
			ipv6Routes = append(ipv6Routes, fmt.Sprintf("%s ::", r))
		} else {
			ipv4Routes = append(ipv4Routes, fmt.Sprintf("%s 0.0.0.0", r))
		}
	}

	// Agregar IPs IPv4 de dominios resueltos como /32
	for _, ip := range targets.IPv4 {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			ipv4Routes = append(ipv4Routes, fmt.Sprintf("%s/32 0.0.0.0", ip))
		}
	}

	isDual := IsHostDualStack()
	vpnHasIPv6 := m.hasIPv6Support(conn.Name)

	args := []string{"connection", "modify", conn.Name,
		"ipv4.routes", strings.Join(ipv4Routes, ", "),
		"ipv4.never-default", "yes",
		"ipv4.ignore-auto-routes", "yes",
	}

	if isDual && vpnHasIPv6 {
		// Dual-Stack Completo: Tanto el host local como la VPN soportan IPv6
		for _, ip := range targets.IPv6 {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				ipv6Routes = append(ipv6Routes, fmt.Sprintf("%s/128 ::", ip))
			}
		}

		args = append(args,
			"ipv6.routes", strings.Join(ipv6Routes, ", "),
			"ipv6.never-default", "yes",
			"ipv6.ignore-auto-routes", "yes",
		)
		m.clearBlackholeIPv6(conn.Name)
	} else if isDual && !vpnHasIPv6 {
		// Host tiene IPv6 pero VPN es solo IPv4 -> Opción A: Blindaje Activo Anti-Fugas
		args = append(args,
			"ipv6.method", "disabled",
			"ipv6.routes", "",
			"ipv6.addresses", "",
			"ipv6.gateway", "",
		)
		if len(targets.IPv6) > 0 {
			m.applyBlackholeIPv6(conn.Name, targets.IPv6)
		}
	} else {
		// Host solo IPv4
		args = append(args,
			"ipv6.method", "disabled",
			"ipv6.routes", "",
			"ipv6.addresses", "",
			"ipv6.gateway", "",
		)
		m.clearBlackholeIPv6(conn.Name)
	}

	if conn.IsSalida {
		// En perfiles de salida no sobreescribir el DNS del sistema
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

// Down desconecta el perfil y limpia cualquier ruta de blindaje anti-fugas activa
func (m *LinuxManager) Down(connName string) error {
	m.clearBlackholeIPv6(connName)
	cmd := exec.Command("nmcli", "connection", "down", connName)
	_ = cmd.Run()
	return nil
}

// DeleteConnectionProfile elimina el perfil en NetworkManager
func (m *LinuxManager) DeleteConnectionProfile(connName string) error {
	m.clearBlackholeIPv6(connName)
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
