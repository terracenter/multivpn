package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/terracenter/multivpn/internal/models"
	_ "modernc.org/sqlite"
)

// Storage gestiona la conexión SQLite y las operaciones de datos
type Storage struct {
	db       *sql.DB
	dbPath   string
	dumpPath string
	mu       sync.Mutex
}

// NewStorage inicializa la base de datos SQLite y carga datos si está vacía
func NewStorage(dbPath, dumpPath string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de base de datos: %w", err)
	}

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo base de datos SQLite: %w", err)
	}

	// PRAGMA de rendimiento y llaves foráneas
	if _, err := database.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("error configurando pragmas sqlite: %w", err)
	}

	s := &Storage{
		db:       database,
		dbPath:   dbPath,
		dumpPath: dumpPath,
	}

	if err := s.initSchema(); err != nil {
		_ = database.Close()
		return nil, err
	}

	// Verificar si hay datos; si está vacía e importFile existe, importar
	var count int
	_ = database.QueryRow("SELECT COUNT(*) FROM vpn_connections").Scan(&count)
	if count == 0 && dumpPath != "" {
		if _, err := os.Stat(dumpPath); err == nil {
			if err := s.importDump(dumpPath); err != nil {
				fmt.Printf("[ADVERTENCIA] No se pudo importar dump inicial: %v\n", err)
			}
		}
	}

	return s, nil
}

// Close cierra la conexión de la base de datos
func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS vpn_connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		empresa TEXT,
		is_salida INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS vpn_routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		vpn_id INTEGER NOT NULL,
		route TEXT NOT NULL,
		description TEXT,
		FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,
		UNIQUE(vpn_id, route)
	);
	CREATE TABLE IF NOT EXISTS vpn_domains (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		vpn_id INTEGER NOT NULL,
		domain TEXT NOT NULL,
		description TEXT,
		FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,
		UNIQUE(vpn_id, domain)
	);`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) importDump(dumpFile string) error {
	content, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}

	queries := string(content)
	// Normalizar CREATE TABLE para evitar colisión si initSchema ya se ejecutó
	queries = strings.ReplaceAll(queries, "CREATE TABLE IF NOT EXISTS ", "CREATE TABLE ")
	queries = strings.ReplaceAll(queries, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ")

	// Ejecutar script SQL de volcado
	_, err = s.db.Exec(queries)
	return err
}

// DumpSQL exporta la base de datos completa a un archivo SQL para control de versiones en Git
func (s *Storage) DumpSQL() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dumpPath == "" {
		return nil
	}

	_ = os.MkdirAll(filepath.Dir(s.dumpPath), 0755)

	var sb strings.Builder
	sb.WriteString("PRAGMA foreign_keys=OFF;\nBEGIN TRANSACTION;\n")
	sb.WriteString("CREATE TABLE IF NOT EXISTS vpn_connections (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n    name TEXT UNIQUE NOT NULL,\n    type TEXT NOT NULL,\n    empresa TEXT,\n    is_salida INTEGER DEFAULT 0\n);\n")

	// Dump Connections
	rows, err := s.db.Query("SELECT id, name, type, COALESCE(empresa, ''), is_salida FROM vpn_connections ORDER BY id")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var name, connType, empresa string
			var isSalida int
			if err := rows.Scan(&id, &name, &connType, &empresa, &isSalida); err == nil {
				sb.WriteString(fmt.Sprintf("INSERT OR REPLACE INTO vpn_connections VALUES(%d,'%s','%s','%s',%d);\n",
					id, strings.ReplaceAll(name, "'", "''"), connType, strings.ReplaceAll(empresa, "'", "''"), isSalida))
			}
		}
	}

	sb.WriteString("CREATE TABLE IF NOT EXISTS vpn_routes (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n    vpn_id INTEGER NOT NULL,\n    route TEXT NOT NULL,\n    description TEXT,\n    FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,\n    UNIQUE(vpn_id, route)\n);\n")

	// Dump Routes
	rRows, err := s.db.Query("SELECT id, vpn_id, route, COALESCE(description, '') FROM vpn_routes ORDER BY id")
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var id, vpnID int64
			var route, desc string
			if err := rRows.Scan(&id, &vpnID, &route, &desc); err == nil {
				sb.WriteString(fmt.Sprintf("INSERT OR REPLACE INTO vpn_routes VALUES(%d,%d,'%s','%s');\n",
					id, vpnID, strings.ReplaceAll(route, "'", "''"), strings.ReplaceAll(desc, "'", "''")))
			}
		}
	}

	sb.WriteString("CREATE TABLE IF NOT EXISTS vpn_domains (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n    vpn_id INTEGER NOT NULL,\n    domain TEXT NOT NULL,\n    description TEXT,\n    FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,\n    UNIQUE(vpn_id, domain)\n);\n")

	// Dump Domains
	dRows, err := s.db.Query("SELECT id, vpn_id, domain, COALESCE(description, '') FROM vpn_domains ORDER BY id")
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var id, vpnID int64
			var domain, desc string
			if err := dRows.Scan(&id, &vpnID, &domain, &desc); err == nil {
				sb.WriteString(fmt.Sprintf("INSERT OR REPLACE INTO vpn_domains VALUES(%d,%d,'%s','%s');\n",
					id, vpnID, strings.ReplaceAll(domain, "'", "''"), strings.ReplaceAll(desc, "'", "''")))
			}
		}
	}

	sb.WriteString("COMMIT;\n")

	return os.WriteFile(s.dumpPath, []byte(sb.String()), 0644)
}

// GetConnections retorna todas las VPNs registradas
func (s *Storage) GetConnections() ([]models.Connection, error) {
	rows, err := s.db.Query("SELECT id, name, type, COALESCE(empresa, ''), is_salida FROM vpn_connections ORDER BY is_salida ASC, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Connection
	for rows.Next() {
		var c models.Connection
		var isSalida int
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Empresa, &isSalida); err != nil {
			return nil, err
		}
		c.IsSalida = isSalida == 1
		list = append(list, c)
	}
	return list, nil
}

// GetConnection obtiene una VPN por ID
func (s *Storage) GetConnection(id int64) (*models.Connection, error) {
	var c models.Connection
	var isSalida int
	err := s.db.QueryRow("SELECT id, name, type, COALESCE(empresa, ''), is_salida FROM vpn_connections WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Type, &c.Empresa, &isSalida)
	if err != nil {
		return nil, err
	}
	c.IsSalida = isSalida == 1
	return &c, nil
}

// UpsertConnection registra o actualiza una conexión VPN
func (s *Storage) UpsertConnection(name, connType, empresa string, isSalida bool) (int64, error) {
	salidaVal := 0
	if isSalida {
		salidaVal = 1
	}

	res, err := s.db.Exec(`
		INSERT INTO vpn_connections (name, type, empresa, is_salida)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			type = excluded.type,
			empresa = excluded.empresa,
			is_salida = excluded.is_salida`,
		name, connType, empresa, salidaVal)
	if err != nil {
		return 0, err
	}
	_ = s.DumpSQL()
	return res.LastInsertId()
}

// DeleteConnection elimina una conexión VPN y sus rutas y dominios en cascada
func (s *Storage) DeleteConnection(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM vpn_routes WHERE vpn_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM vpn_domains WHERE vpn_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM vpn_connections WHERE id = ?", id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	_ = s.DumpSQL()
	return nil
}

// ReassignConnectionData reasigna todas las rutas y dominios de fromVPNID a toVPNID con deduplicación estricta y elimina fromVPNID
func (s *Storage) ReassignConnectionData(fromVPNID, toVPNID int64) error {
	if fromVPNID == toVPNID {
		return fmt.Errorf("la VPN de origen y destino no pueden ser la misma")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Reasignar rutas aplicando deduplicación (INSERT OR IGNORE)
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO vpn_routes (vpn_id, route, description)
		SELECT ?, route, description FROM vpn_routes WHERE vpn_id = ?`, toVPNID, fromVPNID)
	if err != nil {
		return fmt.Errorf("error reasignando rutas: %w", err)
	}

	// Borrar las rutas originales de la VPN de origen
	if _, err = tx.Exec("DELETE FROM vpn_routes WHERE vpn_id = ?", fromVPNID); err != nil {
		return err
	}

	// 2. Reasignar dominios aplicando deduplicación (INSERT OR IGNORE)
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO vpn_domains (vpn_id, domain, description)
		SELECT ?, domain, description FROM vpn_domains WHERE vpn_id = ?`, toVPNID, fromVPNID)
	if err != nil {
		return fmt.Errorf("error reasignando dominios: %w", err)
	}

	// Borrar los dominios originales de la VPN de origen
	if _, err = tx.Exec("DELETE FROM vpn_domains WHERE vpn_id = ?", fromVPNID); err != nil {
		return err
	}

	// 3. Eliminar la VPN de origen
	if _, err = tx.Exec("DELETE FROM vpn_connections WHERE id = ?", fromVPNID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	_ = s.DumpSQL()
	return nil
}

// GetRoutes retorna las rutas asociadas, con soporte opcional de filtro por vpn_id y búsqueda textual
func (s *Storage) GetRoutes(vpnID int64, search string) ([]models.Route, error) {
	query := `
		SELECT r.id, r.vpn_id, c.name, r.route, COALESCE(r.description, '')
		FROM vpn_routes r
		JOIN vpn_connections c ON r.vpn_id = c.id
		WHERE 1=1`
	var args []interface{}

	if vpnID > 0 {
		query += " AND r.vpn_id = ?"
		args = append(args, vpnID)
	}

	if search != "" {
		query += " AND (r.route LIKE ? OR r.description LIKE ? OR c.name LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query += " ORDER BY c.name ASC, r.route ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Route
	for rows.Next() {
		var r models.Route
		if err := rows.Scan(&r.ID, &r.VPNID, &r.VPNName, &r.Route, &r.Description); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// GetRoute obtiene una ruta específica por ID
func (s *Storage) GetRoute(id int64) (*models.Route, error) {
	var r models.Route
	err := s.db.QueryRow(`
		SELECT r.id, r.vpn_id, c.name, r.route, COALESCE(r.description, '')
		FROM vpn_routes r
		JOIN vpn_connections c ON r.vpn_id = c.id
		WHERE r.id = ?`, id).
		Scan(&r.ID, &r.VPNID, &r.VPNName, &r.Route, &r.Description)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// AddRoute agrega una nueva ruta IP/CIDR
func (s *Storage) AddRoute(vpnID int64, route, description string) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO vpn_routes (vpn_id, route, description) VALUES (?, ?, ?)",
		vpnID, strings.TrimSpace(route), strings.TrimSpace(description))
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// UpdateRoute MODIFICA una ruta existente (funcionalidad clave para edición amigable)
func (s *Storage) UpdateRoute(id int64, vpnID int64, route, description string) error {
	_, err := s.db.Exec("UPDATE vpn_routes SET vpn_id = ?, route = ?, description = ? WHERE id = ?",
		vpnID, strings.TrimSpace(route), strings.TrimSpace(description), id)
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// DeleteRoute elimina una ruta por ID
func (s *Storage) DeleteRoute(id int64) error {
	_, err := s.db.Exec("DELETE FROM vpn_routes WHERE id = ?", id)
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// GetDomains retorna los dominios registrados
func (s *Storage) GetDomains(vpnID int64, search string) ([]models.Domain, error) {
	query := `
		SELECT d.id, d.vpn_id, c.name, d.domain, COALESCE(d.description, '')
		FROM vpn_domains d
		JOIN vpn_connections c ON d.vpn_id = c.id
		WHERE 1=1`
	var args []interface{}

	if vpnID > 0 {
		query += " AND d.vpn_id = ?"
		args = append(args, vpnID)
	}

	if search != "" {
		query += " AND (d.domain LIKE ? OR d.description LIKE ? OR c.name LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query += " ORDER BY c.name ASC, d.domain ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Domain
	for rows.Next() {
		var d models.Domain
		if err := rows.Scan(&d.ID, &d.VPNID, &d.VPNName, &d.Domain, &d.Description); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

// GetDomain obtiene un dominio por ID
func (s *Storage) GetDomain(id int64) (*models.Domain, error) {
	var d models.Domain
	err := s.db.QueryRow(`
		SELECT d.id, d.vpn_id, c.name, d.domain, COALESCE(d.description, '')
		FROM vpn_domains d
		JOIN vpn_connections c ON d.vpn_id = c.id
		WHERE d.id = ?`, id).
		Scan(&d.ID, &d.VPNID, &d.VPNName, &d.Domain, &d.Description)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// AddDomain agrega un nuevo dominio
func (s *Storage) AddDomain(vpnID int64, domain, description string) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO vpn_domains (vpn_id, domain, description) VALUES (?, ?, ?)",
		vpnID, strings.TrimSpace(domain), strings.TrimSpace(description))
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// UpdateDomain MODIFICA un dominio existente
func (s *Storage) UpdateDomain(id int64, vpnID int64, domain, description string) error {
	_, err := s.db.Exec("UPDATE vpn_domains SET vpn_id = ?, domain = ?, description = ? WHERE id = ?",
		vpnID, strings.TrimSpace(domain), strings.TrimSpace(description), id)
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// DeleteDomain elimina un dominio por ID
func (s *Storage) DeleteDomain(id int64) error {
	_, err := s.db.Exec("DELETE FROM vpn_domains WHERE id = ?", id)
	if err == nil {
		_ = s.DumpSQL()
	}
	return err
}

// GlobalSearch realiza una búsqueda unificada por nombre de VPN, empresa, IP/red CIDR y dominios
func (s *Storage) GlobalSearch(query string) (*models.SearchResults, error) {
	query = strings.TrimSpace(query)
	results := &models.SearchResults{
		Query:       query,
		Connections: []models.Connection{},
		Routes:      []models.Route{},
		Domains:     []models.Domain{},
	}

	if query == "" {
		return results, nil
	}

	pattern := "%" + query + "%"

	// 1. Buscar en conexiones
	cRows, err := s.db.Query(`
		SELECT id, name, type, COALESCE(empresa, ''), is_salida
		FROM vpn_connections
		WHERE name LIKE ? OR empresa LIKE ?
		ORDER BY name ASC`, pattern, pattern)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var c models.Connection
			var isSalida int
			if err := cRows.Scan(&c.ID, &c.Name, &c.Type, &c.Empresa, &isSalida); err == nil {
				c.IsSalida = isSalida == 1
				results.Connections = append(results.Connections, c)
			}
		}
	}

	// 2. Buscar en rutas
	rRows, err := s.db.Query(`
		SELECT r.id, r.vpn_id, c.name, r.route, COALESCE(r.description, '')
		FROM vpn_routes r
		JOIN vpn_connections c ON r.vpn_id = c.id
		WHERE r.route LIKE ? OR r.description LIKE ? OR c.name LIKE ?
		ORDER BY c.name ASC, r.route ASC`, pattern, pattern, pattern)
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var r models.Route
			if err := rRows.Scan(&r.ID, &r.VPNID, &r.VPNName, &r.Route, &r.Description); err == nil {
				results.Routes = append(results.Routes, r)
			}
		}
	}

	// 3. Buscar en dominios
	dRows, err := s.db.Query(`
		SELECT d.id, d.vpn_id, c.name, d.domain, COALESCE(d.description, '')
		FROM vpn_domains d
		JOIN vpn_connections c ON d.vpn_id = c.id
		WHERE d.domain LIKE ? OR d.description LIKE ? OR c.name LIKE ?
		ORDER BY c.name ASC, d.domain ASC`, pattern, pattern, pattern)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var d models.Domain
			if err := dRows.Scan(&d.ID, &d.VPNID, &d.VPNName, &d.Domain, &d.Description); err == nil {
				results.Domains = append(results.Domains, d)
			}
		}
	}

	results.TotalCount = len(results.Connections) + len(results.Routes) + len(results.Domains)
	return results, nil
}

// GetStats calcula las métricas globales para el Dashboard
func (s *Storage) GetStats() (*models.Stats, error) {
	var stats models.Stats
	_ = s.db.QueryRow("SELECT COUNT(*) FROM vpn_connections").Scan(&stats.TotalVPNs)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM vpn_routes").Scan(&stats.TotalRoutes)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM vpn_domains").Scan(&stats.TotalDomains)
	return &stats, nil
}
