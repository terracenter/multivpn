package network

import (
	"context"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/terracenter/multivpn/internal/models"
)

// ResolveDomainsParallel resuelve concurrentemente una lista de nombres de dominio clasificándolos en IPv4 e IPv6
func ResolveDomainsParallel(domains []string, timeout time.Duration, isDualStack bool) models.ResolvedTargets {
	var targets models.ResolvedTargets
	if len(domains) == 0 {
		return targets
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ipv4Set := make(map[string]struct{})
	ipv6Set := make(map[string]struct{})

	networkType := "ip4"
	if isDualStack {
		networkType = "ip"
	}

	for _, rawDomain := range domains {
		clean := cleanDomain(rawDomain)
		if clean == "" {
			continue
		}

		wg.Add(1)
		go func(dom string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			ips, err := net.DefaultResolver.LookupIP(ctx, networkType, dom)
			if err != nil {
				return
			}

			mu.Lock()
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					ipv4Set[ipv4.String()] = struct{}{}
				} else if isDualStack && ip.To16() != nil {
					ipv6Set[ip.String()] = struct{}{}
				}
			}
			mu.Unlock()
		}(clean)
	}

	wg.Wait()

	for ip := range ipv4Set {
		targets.IPv4 = append(targets.IPv4, ip)
	}
	for ip := range ipv6Set {
		targets.IPv6 = append(targets.IPv6, ip)
	}

	return targets
}

// cleanDomain limpia URLs o paths dejando únicamente el hostname FQDN
func cleanDomain(raw string) string {
	d := strings.TrimSpace(raw)
	if strings.Contains(d, "://") {
		if u, err := url.Parse(d); err == nil {
			d = u.Hostname()
		}
	} else if strings.Contains(d, "/") {
		parts := strings.Split(d, "/")
		d = parts[0]
	}
	return strings.TrimSpace(d)
}
