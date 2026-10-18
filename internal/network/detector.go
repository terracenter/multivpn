package network

import (
	"net"
	"time"
)

// IsHostDualStack determina si el host local tiene conectividad IPv6 enrutable hacia internet
func IsHostDualStack() bool {
	// Comprobar si hay una ruta válida IPv6 en el kernel intentando un Dial UDP rápido hacia Cloudflare DNS IPv6.
	// En UDP, Dial no envía paquetes físicos al exterior, solo consulta la tabla de enrutamiento del kernel local.
	conn, err := net.DialTimeout("udp6", "[2606:4700:4700::1111]:53", 500*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return true
	}

	// Segundo intento de respaldo con Google DNS IPv6
	connGoogle, errGoogle := net.DialTimeout("udp6", "[2001:4860:4860::8888]:53", 500*time.Millisecond)
	if errGoogle == nil {
		_ = connGoogle.Close()
		return true
	}

	return false
}
