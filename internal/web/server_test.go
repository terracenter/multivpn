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

	// Test GET /api/search/quick con q vacío
	reqEmpty, _ := http.NewRequest(http.MethodGet, "/api/search/quick", nil)
	rrEmpty := httptest.NewRecorder()
	server.handleSearchQuick(rrEmpty, reqEmpty)

	if rrEmpty.Code != http.StatusOK {
		t.Errorf("Código de estado esperado 200 para /api/search/quick vacío, se obtuvo %d", rrEmpty.Code)
	}
}

