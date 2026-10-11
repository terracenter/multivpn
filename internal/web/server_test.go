package web

import (
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
}
