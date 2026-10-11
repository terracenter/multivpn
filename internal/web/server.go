package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/terracenter/multivpn/internal/config"
	"github.com/terracenter/multivpn/internal/db"
	"github.com/terracenter/multivpn/internal/network"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templateFS embed.FS

// Server es el servidor web embebido para la interfaz interactiva
type Server struct {
	cfg       *config.Config
	store     *db.Storage
	netMgr    network.VPNProvider
	templates map[string]*template.Template
	version   string
}

// NewServer inicializa los manejadores y plantillas embebidas
func NewServer(cfg *config.Config, store *db.Storage, netMgr network.VPNProvider, version string) (*Server, error) {
	s := &Server{
		cfg:       cfg,
		store:     store,
		netMgr:    netMgr,
		templates: make(map[string]*template.Template),
		version:   version,
	}

	if err := s.parseTemplates(); err != nil {
		return nil, fmt.Errorf("error analizando plantillas HTML: %w", err)
	}

	return s, nil
}

func (s *Server) parseTemplates() error {
	// Parsear Dashboard
	tDash, err := template.ParseFS(templateFS, "templates/layout.html", "templates/dashboard.html", "templates/vpn_card.html")
	if err != nil {
		return err
	}
	s.templates["dashboard"] = tDash

	// Parsear Tarjeta individual de VPN (para swaps HTMX)
	tCard, err := template.ParseFS(templateFS, "templates/vpn_card.html")
	if err != nil {
		return err
	}
	s.templates["vpn_card"] = tCard

	// Parsear Rutas
	tRoutes, err := template.ParseFS(templateFS, "templates/layout.html", "templates/routes.html", "templates/route_rows.html")
	if err != nil {
		return err
	}
	s.templates["routes"] = tRoutes

	// Parsear Filas de Rutas (para respuestas parciales HTMX)
	tRouteRows, err := template.ParseFS(templateFS, "templates/route_rows.html")
	if err != nil {
		return err
	}
	s.templates["route_rows"] = tRouteRows

	// Parsear Dominios
	tDomains, err := template.ParseFS(templateFS, "templates/layout.html", "templates/domains.html", "templates/domain_rows.html")
	if err != nil {
		return err
	}
	s.templates["domains"] = tDomains

	// Parsear Filas de Dominios (para respuestas parciales HTMX)
	tDomainRows, err := template.ParseFS(templateFS, "templates/domain_rows.html")
	if err != nil {
		return err
	}
	s.templates["domain_rows"] = tDomainRows

	// Parsear Búsqueda Completa
	tSearch, err := template.ParseFS(templateFS, "templates/layout.html", "templates/search.html")
	if err != nil {
		return err
	}
	s.templates["search"] = tSearch

	// Parsear Búsqueda Rápida (Modal)
	tSearchQuick, err := template.ParseFS(templateFS, "templates/search_quick.html")
	if err != nil {
		return err
	}
	s.templates["search_quick"] = tSearchQuick

	return nil
}

// Start levanta el servidor HTTP local
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Servir archivos estáticos embebidos
	subStatic, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subStatic))))

	// Vistas Web
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/routes", s.handleRoutes)
	mux.HandleFunc("/domains", s.handleDomains)
	mux.HandleFunc("/search", s.handleSearchPage)

	// API Búsqueda
	mux.HandleFunc("/api/search/quick", s.handleSearchQuick)

	// API VPNs
	mux.HandleFunc("/api/vpns/up", s.handleVPNUp)
	mux.HandleFunc("/api/vpns/down", s.handleVPNDown)
	mux.HandleFunc("/api/vpns/shield", s.handleVPNShield)
	mux.HandleFunc("/api/vpns/shield-all", s.handleVPNShieldAll)
	mux.HandleFunc("/api/vpns/up-all", s.handleVPNUpAll)
	mux.HandleFunc("/api/vpns/down-all", s.handleVPNDownAll)
	mux.HandleFunc("/api/vpns/scan", s.handleVPNScan)
	mux.HandleFunc("/api/vpns/reassign", s.handleVPNReassign)
	mux.HandleFunc("/api/vpns/delete", s.handleVPNDelete)
	mux.HandleFunc("/api/vpns/", s.handleVPNItem)

	// API Rutas
	mux.HandleFunc("/api/routes", s.handleRoutesCollection)
	mux.HandleFunc("/api/routes/search", s.handleRoutesSearch)
	mux.HandleFunc("/api/routes/", s.handleRouteItem)

	// API Dominios
	mux.HandleFunc("/api/domains", s.handleDomainsCollection)
	mux.HandleFunc("/api/domains/search", s.handleDomainsSearch)
	mux.HandleFunc("/api/domains/", s.handleDomainItem)

	addr := fmt.Sprintf("%s:%s", s.cfg.WebHost, s.cfg.WebPort)
	fmt.Printf("[MultiVPN WebUI] Servidor listo en http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}
