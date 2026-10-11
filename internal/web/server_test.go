package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/terracenter/multivpn/internal/config"
	"github.com/terracenter/multivpn/internal/db"
	"github.com/terracenter/multivpn/internal/network"
)

func TestServerTemplateParsing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "web_test_*")
	if err != nil {
		t.Fatalf("Error creando temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	dumpCopy := filepath.Join(tempDir, "test_dump.sql")
	origDump := filepath.Join("..", "..", "data", "vpn_data_dump.sql")
	content, _ := os.ReadFile(origDump)
	_ = os.WriteFile(dumpCopy, content, 0644)

	store, err := db.NewStorage(dbPath, dumpCopy)
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		ConfigDir:   tempDir,
		DBPath:      dbPath,
		DumpSQLPath: dumpCopy,
		WebPort:     "8765",
		WebHost:     "127.0.0.1",
	}

	netMgr := network.NewLinuxManager()
	server, err := NewServer(cfg, store, netMgr, "1.0.0-test")
	if err != nil {
		t.Fatalf("Error inicializando Server: %v", err)
	}

	if len(server.templates) == 0 {
		t.Errorf("Se esperaba que los templates estuvieran parseados")
	}

	if _, ok := server.templates["search"]; !ok {
		t.Errorf("Plantilla 'search' no fue parseada")
	}

	if _, ok := server.templates["search_quick"]; !ok {
		t.Errorf("Plantilla 'search_quick' no fue parseada")
	}
}

func TestSearchEndpoints(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "web_search_test_*")
	if err != nil {
		t.Fatalf("Error creando temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	dumpCopy := filepath.Join(tempDir, "test_dump.sql")
	origDump := filepath.Join("..", "..", "data", "vpn_data_dump.sql")
	content, _ := os.ReadFile(origDump)
	_ = os.WriteFile(dumpCopy, content, 0644)

	store, err := db.NewStorage(dbPath, dumpCopy)
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		ConfigDir:   tempDir,
		DBPath:      dbPath,
		DumpSQLPath: dumpCopy,
		WebPort:     "8765",
		WebHost:     "127.0.0.1",
	}

	netMgr := network.NewLinuxManager()
	server, err := NewServer(cfg, store, netMgr, "1.0.0-test")
	if err != nil {
		t.Fatalf("Error inicializando Server: %v", err)
	}

	// Test GET /search?q=10.
	req, _ := http.NewRequest(http.MethodGet, "/search?q=10.", nil)
	rr := httptest.NewRecorder()
	server.handleSearchPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /search, se obtuvo %d", rr.Code)
	}

	// Test GET /api/search/quick?q=10.
	reqQuick, _ := http.NewRequest(http.MethodGet, "/api/search/quick?q=10.", nil)
	rrQuick := httptest.NewRecorder()
	server.handleSearchQuick(rrQuick, reqQuick)

	if rrQuick.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/search/quick, se obtuvo %d", rrQuick.Code)
	}

	// Test GET /api/routes/search con filtros de columna
	reqRouteFilter, _ := http.NewRequest(http.MethodGet, "/api/routes/search?route=10.&description=NOC", nil)
	rrRouteFilter := httptest.NewRecorder()
	server.handleRoutesSearch(rrRouteFilter, reqRouteFilter)
	if rrRouteFilter.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/routes/search, se obtuvo %d", rrRouteFilter.Code)
	}

	// Test GET /api/domains/search con filtros de columna
	reqDomainFilter, _ := http.NewRequest(http.MethodGet, "/api/domains/search?domain=claude&description=AI", nil)
	rrDomainFilter := httptest.NewRecorder()
	server.handleDomainsSearch(rrDomainFilter, reqDomainFilter)
	if rrDomainFilter.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/domains/search, se obtuvo %d", rrDomainFilter.Code)
	}

	// Test GET /api/routes/search con ordenamiento por route DESC
	reqRouteSort, _ := http.NewRequest(http.MethodGet, "/api/routes/search?sort_by=route&sort_order=desc", nil)
	rrRouteSort := httptest.NewRecorder()
	server.handleRoutesSearch(rrRouteSort, reqRouteSort)
	if rrRouteSort.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/routes/search con ordenamiento, se obtuvo %d", rrRouteSort.Code)
	}

	// Test GET /api/domains/search con ordenamiento por domain ASC
	reqDomainSort, _ := http.NewRequest(http.MethodGet, "/api/domains/search?sort_by=domain&sort_order=asc", nil)
	rrDomainSort := httptest.NewRecorder()
	server.handleDomainsSearch(rrDomainSort, reqDomainSort)
	if rrDomainSort.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/domains/search con ordenamiento, se obtuvo %d", rrDomainSort.Code)
	}
}

