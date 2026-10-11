package models

// Connection representa una conexión VPN registrada
type Connection struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Type               string `json:"type"` // "vpn" (L2TP/PPTP) o "wireguard"
	Empresa            string `json:"empresa"`
	IsSalida           bool   `json:"is_salida"` // 1 si es salida de sitios bloqueados
	IsActive           bool   `json:"is_active"` // Estado en tiempo real
	HasDefaultGateway  bool   `json:"has_default_gateway"` // true si tiene riesgo de fuga (never-default=no)
}

// Route representa una ruta estática (IP/CIDR) asociada a una VPN
type Route struct {
	ID          int64  `json:"id"`
	VPNID       int64  `json:"vpn_id"`
	VPNName     string `json:"vpn_name,omitempty"`
	Route       string `json:"route"`
	Description string `json:"description"`
}

// Domain representa un dominio web asociado a una VPN de salida
type Domain struct {
	ID          int64  `json:"id"`
	VPNID       int64  `json:"vpn_id"`
	VPNName     string `json:"vpn_name,omitempty"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

// Stats contiene las métricas para el Dashboard
type Stats struct {
	TotalVPNs    int `json:"total_vpns"`
	ActiveVPNs   int `json:"active_vpns"`
	TotalRoutes  int `json:"total_routes"`
	TotalDomains int `json:"total_domains"`
	ShieldedVPNs int `json:"shielded_vpns"`
}

// SearchResults contiene los resultados consolidados de una búsqueda global
type SearchResults struct {
	Query       string       `json:"query"`
	Connections []Connection `json:"connections"`
	Routes      []Route      `json:"routes"`
	Domains     []Domain     `json:"domains"`
	TotalCount  int          `json:"total_count"`
}
