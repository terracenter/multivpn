package network

import (
	"testing"
	"time"
)

func TestCleanDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://claude.ai/chat", "claude.ai"},
		{"http://test.com", "test.com"},
		{"api.brave.com/search", "api.brave.com"},
		{"  google.com  ", "google.com"},
		{"sub.domain.org", "sub.domain.org"},
	}

	for _, tt := range tests {
		got := cleanDomain(tt.input)
		if got != tt.expected {
			t.Errorf("cleanDomain(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestResolveDomainsParallel(t *testing.T) {
	// Probar con un dominio estándar de DNS
	domains := []string{"one.one.one.one", "dns.google"}
	ips := ResolveDomainsParallel(domains, 3*time.Second)

	if len(ips) == 0 {
		t.Logf("Aviso: no se resolvieron IPs (posible falta de conexión externa en sandbox)")
	} else {
		t.Logf("IPs resueltas con éxito en Go nativo: %v", ips)
	}
}
