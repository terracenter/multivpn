package network

import (
	"context"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ResolveDomainsParallel resuelve concurrentemente una lista de nombres de dominio a direcciones IPv4
func ResolveDomainsParallel(domains []string, timeout time.Duration) []string {
	if len(domains) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ipSet := make(map[string]struct{})

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

			ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", dom)
			if err != nil {
				return
			}

			mu.Lock()
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					ipSet[ipv4.String()] = struct{}{}
				}
			}
			mu.Unlock()
		}(clean)
	}

	wg.Wait()

	result := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		result = append(result, ip)
	}

	return result
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
