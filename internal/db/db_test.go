package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageImportAndCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "multivpn_test_*")
	if err != nil {
		t.Fatalf("Error creando temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	dumpCopy := filepath.Join(tempDir, "test_dump.sql")

	// Usar el dump real del proyecto
	origDump := filepath.Join("..", "..", "data", "vpn_data_dump.sql")
	content, err := os.ReadFile(origDump)
	if err != nil {
		t.Fatalf("No se pudo leer data/vpn_data_dump.sql: %v", err)
	}
	if err := os.WriteFile(dumpCopy, content, 0644); err != nil {
		t.Fatalf("No se pudo copiar dump: %v", err)
	}

	store, err := NewStorage(dbPath, dumpCopy)
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer store.Close()

	// 1. Verificar importación
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("Error obteniendo stats: %v", err)
	}

	if stats.TotalVPNs == 0 {
		t.Errorf("Se esperaban VPNs importadas, se obtuvieron 0")
	}
	if stats.TotalRoutes < 60 {
		t.Errorf("Se esperaban al menos 60 rutas, se obtuvieron %d", stats.TotalRoutes)
	}
	if stats.TotalDomains < 30 {
		t.Errorf("Se esperaban al menos 30 dominios, se obtuvieron %d", stats.TotalDomains)
	}

	// 2. Probar Creación de Ruta
	conns, _ := store.GetConnections()
	if len(conns) == 0 {
		t.Fatalf("No hay conexiones disponibles para prueba")
	}
	testVPN := conns[0].ID

	err = store.AddRoute(testVPN, "10.99.99.0/24", "Red de Prueba Inicial")
	if err != nil {
		t.Fatalf("Error agregando ruta: %v", err)
	}

	// Buscar ruta
	routes, err := store.GetRoutes(testVPN, "10.99.99.0")
	if err != nil || len(routes) == 0 {
		t.Fatalf("No se encontró la ruta creada: %v", err)
	}
	createdID := routes[0].ID

	// 3. Probar Modificación / Edición directa (la funcionalidad solicitada)
	err = store.UpdateRoute(createdID, testVPN, "10.99.99.0/24", "Descripción Modificada Exitosamente")
	if err != nil {
		t.Fatalf("Error modificando ruta: %v", err)
	}

	updatedRoute, err := store.GetRoute(createdID)
	if err != nil {
		t.Fatalf("Error obteniendo ruta modificada: %v", err)
	}
	if updatedRoute.Description != "Descripción Modificada Exitosamente" {
		t.Errorf("La descripción no se actualizó, obtenida: %s", updatedRoute.Description)
	}

	// 4. Probar Eliminación
	err = store.DeleteRoute(createdID)
	if err != nil {
		t.Fatalf("Error eliminando ruta: %v", err)
	}
}

func TestReassignConnectionDataAndDeduplication(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "multivpn_test_reassign_*")
	if err != nil {
		t.Fatalf("Error creando temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	dumpCopy := filepath.Join(tempDir, "test_dump.sql")

	origDump := filepath.Join("..", "..", "data", "vpn_data_dump.sql")
	content, _ := os.ReadFile(origDump)
	_ = os.WriteFile(dumpCopy, content, 0644)

	store, err := NewStorage(dbPath, dumpCopy)
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer store.Close()

	// Crear dos VPNs temporales
	vpn1ID, _ := store.UpsertConnection("TestVPN1", "vpn", "test", false)
	vpn2ID, _ := store.UpsertConnection("TestVPN2", "vpn", "test", true)

	// Agregar rutas a VPN1
	_ = store.AddRoute(vpn1ID, "10.1.1.0/24", "Ruta A")
	_ = store.AddRoute(vpn1ID, "10.2.2.0/24", "Ruta Duplicada Compartida")
	_ = store.AddDomain(vpn1ID, "example.com", "Dominio A")

	// Agregar a VPN2 la misma ruta para probar deduplicación
	_ = store.AddRoute(vpn2ID, "10.2.2.0/24", "Ruta ya existente en destino")

	// Reasignar VPN1 a VPN2
	err = store.ReassignConnectionData(vpn1ID, vpn2ID)
	if err != nil {
		t.Fatalf("Error reasignando datos: %v", err)
	}

	// Verificar que VPN1 fue eliminada
	conns, _ := store.GetConnections()
	for _, c := range conns {
		if c.ID == vpn1ID {
			t.Errorf("VPN1 no debería existir tras reasignación")
		}
	}

	// Verificar que VPN2 tiene las rutas sin duplicados
	routes2, _ := store.GetRoutes(vpn2ID, "")
	countMap := make(map[string]int)
	for _, r := range routes2 {
		countMap[r.Route]++
		if countMap[r.Route] > 1 {
			t.Errorf("Ruta duplicada detectada en destino: %s", r.Route)
		}
	}

	// Probar GlobalSearch
	results, err := store.GlobalSearch("10.1.1.0")
	if err != nil || len(results.Routes) == 0 {
		t.Errorf("GlobalSearch no encontró la ruta 10.1.1.0: %v", err)
	}

	resDomain, err := store.GlobalSearch("example.com")
	if err != nil || len(resDomain.Domains) == 0 {
		t.Errorf("GlobalSearch no encontró el dominio example.com: %v", err)
	}
}
