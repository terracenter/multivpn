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
	domains := []string{"one.one.one.one", "dns.google"}

	// Modo solo IPv4
	targetsV4 := ResolveDomainsParallel(domains, 3*time.Second, false)
	t.Logf("IPs resueltas en modo IPv4: %v", targetsV4.IPv4)
	if len(targetsV4.IPv6) > 0 {
		t.Errorf("Modo IPv4 devolvió IPv6: %v", targetsV4.IPv6)
	}

	// Modo Dual-Stack
	targetsDual := ResolveDomainsParallel(domains, 3*time.Second, true)
	t.Logf("IPs resueltas en modo Dual-Stack -> IPv4: %v, IPv6: %v", targetsDual.IPv4, targetsDual.IPv6)
}
