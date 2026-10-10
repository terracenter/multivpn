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
