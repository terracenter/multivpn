package network

import "github.com/terracenter/multivpn/internal/models"

// VPNProvider define el contrato multiplataforma para gestionar túneles y rutas
type VPNProvider interface {
	// ScanConnections detecta perfiles de VPN existentes en el sistema operativo
	ScanConnections() ([]models.Connection, error)

	// ShieldConnection aplica blindaje de Default Gateway (Split Tunneling estricto)
	ShieldConnection(connName string) error

	// IsConnectionActive verifica si una VPN está conectada en tiempo real
	IsConnectionActive(connName string) (bool, error)

	// HasDefaultGatewayRisk comprueba si la VPN tiene habilitado el gateway remoto
	HasDefaultGatewayRisk(connName string) (bool, error)

	// Up prepara las rutas, aplica la configuración y levanta el túnel
	Up(conn models.Connection, routes []string, targets models.ResolvedTargets) error

	// Down desconecta el túnel especificado
	Down(connName string) error

	// DeleteConnectionProfile elimina el perfil de conexión del gestor del sistema (ej: nmcli)
	DeleteConnectionProfile(connName string) error
}
