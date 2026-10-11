package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/terracenter/multivpn/internal/models"
	"github.com/terracenter/multivpn/internal/network"
)

type VPNFilter struct {
	ID    int64
	Name  string
	Count int
}

type pageData struct {
	Title         string
	ActiveTab     string
	Version       string
	DBPath        string
	Stats         *models.Stats
	Connections   []models.Connection
	Routes        []models.Route
	Domains       []models.Domain
	VPNFilters    []VPNFilter
	SearchResults *models.SearchResults
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	conns, _ := s.store.GetConnections()
	stats, _ := s.store.GetStats()

	// Actualizar estados en vivo
	for i := range conns {
		active, _ := s.netMgr.IsConnectionActive(conns[i].Name)
		hasRisk, _ := s.netMgr.HasDefaultGatewayRisk(conns[i].Name)
		conns[i].IsActive = active
		conns[i].HasDefaultGateway = hasRisk
		if active {
			stats.ActiveVPNs++
		}
	}

	data := pageData{
		Title:       "Dashboard",
		ActiveTab:   "dashboard",
		Version:     s.version,
		DBPath:      s.cfg.DBPath,
		Stats:       stats,
		Connections: conns,
	}

	_ = s.templates["dashboard"].ExecuteTemplate(w, "layout", data)
}

func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	routes, _ := s.store.GetRoutes(0, "")
	conns, _ := s.store.GetConnections()
	stats, _ := s.store.GetStats()

	var filters []VPNFilter
	for _, c := range conns {
		cRoutes, _ := s.store.GetRoutes(c.ID, "")
		filters = append(filters, VPNFilter{
			ID:    c.ID,
			Name:  c.Name,
			Count: len(cRoutes),
		})
	}

	data := pageData{
		Title:       "Rutas e IPs",
		ActiveTab:   "routes",
		Version:     s.version,
		DBPath:      s.cfg.DBPath,
		Stats:       stats,
		Connections: conns,
		Routes:      routes,
		VPNFilters:  filters,
	}

	_ = s.templates["routes"].ExecuteTemplate(w, "layout", data)
}

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	domains, _ := s.store.GetDomains(0, "")
	conns, _ := s.store.GetConnections()
	stats, _ := s.store.GetStats()

	var filters []VPNFilter
	for _, c := range conns {
		cDomains, _ := s.store.GetDomains(c.ID, "")
		filters = append(filters, VPNFilter{
			ID:    c.ID,
			Name:  c.Name,
			Count: len(cDomains),
		})
	}

	data := pageData{
		Title:       "Dominios",
		ActiveTab:   "domains",
		Version:     s.version,
		DBPath:      s.cfg.DBPath,
		Stats:       stats,
		Connections: conns,
		Domains:     domains,
		VPNFilters:  filters,
	}

	_ = s.templates["domains"].ExecuteTemplate(w, "layout", data)
}

// handleVPNUp conecta una VPN y actualiza la tarjeta mediante HTMX
func (s *Server) handleVPNUp(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Nombre requerido", http.StatusBadRequest)
		return
	}

	conns, _ := s.store.GetConnections()
	var target *models.Connection
	for _, c := range conns {
		if c.Name == name {
			target = &c
			break
		}
	}

	if target == nil {
		http.Error(w, "VPN no encontrada", http.StatusNotFound)
		return
	}

	// Obtener rutas estáticas asociadas
	dbRoutes, _ := s.store.GetRoutes(target.ID, "")
	var routeCIDRs []string
	for _, r := range dbRoutes {
		routeCIDRs = append(routeCIDRs, r.Route)
	}

	// Si es salida, resolver dominios en paralelo
	var resolvedIPs []string
	if target.IsSalida {
		dbDomains, _ := s.store.GetDomains(target.ID, "")
		var domainList []string
		for _, d := range dbDomains {
			domainList = append(domainList, d.Domain)
		}
		resolvedIPs = network.ResolveDomainsParallel(domainList, 3*time.Second)
	}

	_ = s.netMgr.Up(*target, routeCIDRs, resolvedIPs)

	// Refrescar estado para la tarjeta
	active, _ := s.netMgr.IsConnectionActive(target.Name)
	hasRisk, _ := s.netMgr.HasDefaultGatewayRisk(target.Name)
	target.IsActive = active
	target.HasDefaultGateway = hasRisk

	_ = s.templates["vpn_card"].ExecuteTemplate(w, "vpn_card", target)
}

// handleVPNDown desconecta una VPN
func (s *Server) handleVPNDown(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Nombre requerido", http.StatusBadRequest)
		return
	}

	_ = s.netMgr.Down(name)

	conns, _ := s.store.GetConnections()
	var target *models.Connection
	for _, c := range conns {
		if c.Name == name {
			target = &c
			break
		}
	}

	if target != nil {
		target.IsActive, _ = s.netMgr.IsConnectionActive(target.Name)
		target.HasDefaultGateway, _ = s.netMgr.HasDefaultGatewayRisk(target.Name)
		_ = s.templates["vpn_card"].ExecuteTemplate(w, "vpn_card", target)
	}
}

// handleVPNShield aplica Split-Tunneling y blindaje de default gateway
func (s *Server) handleVPNShield(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Nombre requerido", http.StatusBadRequest)
		return
	}

	_ = s.netMgr.ShieldConnection(name)

	conns, _ := s.store.GetConnections()
	var target *models.Connection
	for _, c := range conns {
		if c.Name == name {
			target = &c
			break
		}
	}

	if target != nil {
		target.IsActive, _ = s.netMgr.IsConnectionActive(target.Name)
		target.HasDefaultGateway, _ = s.netMgr.HasDefaultGatewayRisk(target.Name)
		_ = s.templates["vpn_card"].ExecuteTemplate(w, "vpn_card", target)
	}
}

// handleVPNShieldAll aplica blindaje a todas las conexiones registradas
func (s *Server) handleVPNShieldAll(w http.ResponseWriter, r *http.Request) {
	conns, _ := s.store.GetConnections()
	for _, c := range conns {
		_ = s.netMgr.ShieldConnection(c.Name)
	}
	w.WriteHeader(http.StatusOK)
}

// handleVPNUpAll conecta todas las VPNs
func (s *Server) handleVPNUpAll(w http.ResponseWriter, r *http.Request) {
	conns, _ := s.store.GetConnections()
	for _, c := range conns {
		dbRoutes, _ := s.store.GetRoutes(c.ID, "")
		var routeCIDRs []string
		for _, rt := range dbRoutes {
			routeCIDRs = append(routeCIDRs, rt.Route)
		}

		var resolvedIPs []string
		if c.IsSalida {
			dbDomains, _ := s.store.GetDomains(c.ID, "")
			var doms []string
			for _, d := range dbDomains {
				doms = append(doms, d.Domain)
			}
			resolvedIPs = network.ResolveDomainsParallel(doms, 3*time.Second)
		}

		_ = s.netMgr.Up(c, routeCIDRs, resolvedIPs)
	}

	// Devolver el fragmento de dashboard actualizado
	s.renderDashboardFragment(w)
}

// handleVPNDownAll desconecta todas las VPNs
func (s *Server) handleVPNDownAll(w http.ResponseWriter, r *http.Request) {
	conns, _ := s.store.GetConnections()
	for _, c := range conns {
		_ = s.netMgr.Down(c.Name)
	}
	s.renderDashboardFragment(w)
}

// handleVPNScan detecta conexiones en NetworkManager y las sincroniza en SQLite
func (s *Server) handleVPNScan(w http.ResponseWriter, r *http.Request) {
	detected, err := s.netMgr.ScanConnections()
	if err == nil {
		for _, d := range detected {
			isSalida := strings.Contains(strings.ToLower(d.Name), "salida") || d.Type == "wireguard"
			empresa := "general"
			if !isSalida {
				empresa = "trabajo"
			}
			_, _ = s.store.UpsertConnection(d.Name, d.Type, empresa, isSalida)
			// Auto-blindar al detectar para proteger el gateway
			_ = s.netMgr.ShieldConnection(d.Name)
		}
	}
	s.renderDashboardFragment(w)
}

// handleVPNDelete elimina una VPN directamente (y opcionalmente en NetworkManager)
func (s *Server) handleVPNDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	idStr := r.FormValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID de VPN no válido", http.StatusBadRequest)
		return
	}

	deleteNM := r.FormValue("delete_nm") == "true" || r.URL.Query().Get("delete_nm") == "true"

	conns, _ := s.store.GetConnections()
	var connName string
	for _, c := range conns {
		if c.ID == id {
			connName = c.Name
			break
		}
	}

	_ = s.store.DeleteConnection(id)

	if deleteNM && connName != "" {
		_ = s.netMgr.DeleteConnectionProfile(connName)
	}

	s.renderDashboardFragment(w)
}

// handleVPNReassign reasigna rutas y dominios de una VPN a otra y luego elimina la VPN de origen
func (s *Server) handleVPNReassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	fromID, _ := strconv.ParseInt(r.FormValue("from_id"), 10, 64)
	toID, _ := strconv.ParseInt(r.FormValue("to_id"), 10, 64)
	deleteNM := r.FormValue("delete_nm") == "true"

	if fromID <= 0 || toID <= 0 {
		http.Error(w, "IDs de VPN no válidos", http.StatusBadRequest)
		return
	}

	conns, _ := s.store.GetConnections()
	var connName string
	for _, c := range conns {
		if c.ID == fromID {
			connName = c.Name
			break
		}
	}

	if err := s.store.ReassignConnectionData(fromID, toID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deleteNM && connName != "" {
		_ = s.netMgr.DeleteConnectionProfile(connName)
	}

	s.renderDashboardFragment(w)
}

// handleVPNItem maneja operaciones sobre /api/vpns/{id}
func (s *Server) handleVPNItem(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/vpns/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID no válido", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		deleteNM := r.URL.Query().Get("delete_nm") == "true"
		conns, _ := s.store.GetConnections()
		var connName string
		for _, c := range conns {
			if c.ID == id {
				connName = c.Name
				break
			}
		}

		_ = s.store.DeleteConnection(id)
		if deleteNM && connName != "" {
			_ = s.netMgr.DeleteConnectionProfile(connName)
		}
		s.renderDashboardFragment(w)
		return
	}

	http.Error(w, "Método no soportado", http.StatusMethodNotAllowed)
}

func (s *Server) renderDashboardFragment(w http.ResponseWriter) {
	conns, _ := s.store.GetConnections()
	stats, _ := s.store.GetStats()
	for i := range conns {
		active, _ := s.netMgr.IsConnectionActive(conns[i].Name)
		hasRisk, _ := s.netMgr.HasDefaultGatewayRisk(conns[i].Name)
		conns[i].IsActive = active
		conns[i].HasDefaultGateway = hasRisk
		if active {
			stats.ActiveVPNs++
		}
	}

	data := pageData{
		Stats:       stats,
		Connections: conns,
	}
	_ = s.templates["dashboard"].ExecuteTemplate(w, "content", data)
}

// handleRoutesSearch responde con filas filtradas para HTMX
func (s *Server) handleRoutesSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	routeFilter := r.URL.Query().Get("route")
	descFilter := r.URL.Query().Get("description")
	if descFilter == "" {
		descFilter = r.URL.Query().Get("desc")
	}
	vpnID, _ := strconv.ParseInt(r.URL.Query().Get("vpn_id"), 10, 64)
	routes, _ := s.store.GetRoutesFiltered(vpnID, routeFilter, descFilter, query)
	data := pageData{Routes: routes}
	_ = s.templates["route_rows"].ExecuteTemplate(w, "route_rows", data)
}

// handleRoutesCollection maneja POST /api/routes para agregar
func (s *Server) handleRoutesCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		vpnID, _ := strconv.ParseInt(r.FormValue("vpn_id"), 10, 64)
		route := r.FormValue("route")
		desc := r.FormValue("description")

		if vpnID > 0 && route != "" {
			_ = s.store.AddRoute(vpnID, route, desc)
		}

		routes, _ := s.store.GetRoutes(0, "")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<tbody id="routes-table-body" class="divide-y divide-slate-800/40">`)
		_ = s.templates["route_rows"].ExecuteTemplate(w, "route_rows", pageData{Routes: routes})
		fmt.Fprint(w, `</tbody>`)
	}
}

// handleRouteItem maneja PUT y DELETE sobre /api/routes/{id}
func (s *Server) handleRouteItem(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID no válido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		_ = r.ParseForm()
		vpnID, _ := strconv.ParseInt(r.FormValue("vpn_id"), 10, 64)
		route := r.FormValue("route")
		desc := r.FormValue("description")

		if vpnID > 0 && route != "" {
			_ = s.store.UpdateRoute(id, vpnID, route, desc)
		}

		routes, _ := s.store.GetRoutes(0, "")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<tbody id="routes-table-body" class="divide-y divide-slate-800/40">`)
		_ = s.templates["route_rows"].ExecuteTemplate(w, "route_rows", pageData{Routes: routes})
		fmt.Fprint(w, `</tbody>`)

	case http.MethodDelete:
		_ = s.store.DeleteRoute(id)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Método no soportado", http.StatusMethodNotAllowed)
	}
}

// handleDomainsSearch responde con filas filtradas para HTMX
func (s *Server) handleDomainsSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	domainFilter := r.URL.Query().Get("domain")
	descFilter := r.URL.Query().Get("description")
	if descFilter == "" {
		descFilter = r.URL.Query().Get("desc")
	}
	vpnID, _ := strconv.ParseInt(r.URL.Query().Get("vpn_id"), 10, 64)
	domains, _ := s.store.GetDomainsFiltered(vpnID, domainFilter, descFilter, query)
	data := pageData{Domains: domains}
	_ = s.templates["domain_rows"].ExecuteTemplate(w, "domain_rows", data)
}

// handleDomainsCollection maneja POST /api/domains
func (s *Server) handleDomainsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		vpnID, _ := strconv.ParseInt(r.FormValue("vpn_id"), 10, 64)
		domain := r.FormValue("domain")
		desc := r.FormValue("description")

		if vpnID > 0 && domain != "" {
			_ = s.store.AddDomain(vpnID, domain, desc)
		}

		domains, _ := s.store.GetDomains(0, "")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<tbody id="domains-table-body" class="divide-y divide-slate-800/40">`)
		_ = s.templates["domain_rows"].ExecuteTemplate(w, "domain_rows", pageData{Domains: domains})
		fmt.Fprint(w, `</tbody>`)
	}
}

// handleDomainItem maneja PUT y DELETE sobre /api/domains/{id}
func (s *Server) handleDomainItem(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/domains/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID no válido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		_ = r.ParseForm()
		vpnID, _ := strconv.ParseInt(r.FormValue("vpn_id"), 10, 64)
		domain := r.FormValue("domain")
		desc := r.FormValue("description")

		if vpnID > 0 && domain != "" {
			_ = s.store.UpdateDomain(id, vpnID, domain, desc)
		}

		domains, _ := s.store.GetDomains(0, "")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<tbody id="domains-table-body" class="divide-y divide-slate-800/40">`)
		_ = s.templates["domain_rows"].ExecuteTemplate(w, "domain_rows", pageData{Domains: domains})
		fmt.Fprint(w, `</tbody>`)

	case http.MethodDelete:
		_ = s.store.DeleteDomain(id)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Método no soportado", http.StatusMethodNotAllowed)
	}
}

// handleSearchPage maneja la vista web completa de resultados de búsqueda (/search?q=...)
func (s *Server) handleSearchPage(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	stats, _ := s.store.GetStats()

	results, err := s.store.GlobalSearch(q)
	if err != nil {
		results = &models.SearchResults{Query: q}
	}

	// Actualizar estado en vivo de VPNs encontradas
	for i := range results.Connections {
		active, _ := s.netMgr.IsConnectionActive(results.Connections[i].Name)
		results.Connections[i].IsActive = active
	}

	data := pageData{
		Title:         "Búsqueda Global",
		ActiveTab:     "search",
		Version:       s.version,
		DBPath:        s.cfg.DBPath,
		Stats:         stats,
		SearchResults: results,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates["search"].ExecuteTemplate(w, "layout", data)
}

// handleSearchQuick maneja la búsqueda reactiva en vivo para el modal flotante (Ctrl+K)
func (s *Server) handleSearchQuick(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<div class="py-6 text-center text-slate-500 text-xs">Escribe para buscar VPNs, rutas CIDR o dominios registrados...</div>`)
		return
	}

	results, err := s.store.GlobalSearch(q)
	if err != nil {
		results = &models.SearchResults{Query: q}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates["search_quick"].ExecuteTemplate(w, "search_quick", pageData{SearchResults: results})
}

