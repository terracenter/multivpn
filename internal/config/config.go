package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Config almacena las configuraciones globales de MultiVPN
type Config struct {
	DBPath      string
	DumpSQLPath string
	WebPort     string
	WebHost     string
}

// LoadConfig lee la configuración local o usa los valores por defecto canónicos
func LoadConfig(baseDir string) *Config {
	defaultDB := "/home/freddy/OneDrive/Sistemas/BasesDatos/vpn_data.db"
	cfgFile := filepath.Join(baseDir, ".vpn_config")
	dumpFile := filepath.Join(baseDir, "data", "vpn_data_dump.sql")

	dbPath := defaultDB

	// Intentar leer .vpn_config si existe
	if file, err := os.Open(cfgFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "DB_PATH=") {
				val := strings.TrimPrefix(line, "DB_PATH=")
				val = strings.Trim(val, `"'`)
				if val != "" {
					dbPath = val
				}
			}
		}
	}

	// Si no existe el directorio de OneDrive, fallback a data/local en el proyecto
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Crear directorio si es posible, o usar fallback
		if err := os.MkdirAll(dir, 0755); err != nil {
			dbPath = filepath.Join(baseDir, "data", "vpn_data.db")
			_ = os.MkdirAll(filepath.Dir(dbPath), 0755)
		}
	}

	port := os.Getenv("MULTIVPN_PORT")
	if port == "" {
		port = "8765"
	}

	host := os.Getenv("MULTIVPN_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	return &Config{
		DBPath:      dbPath,
		DumpSQLPath: dumpFile,
		WebPort:     port,
		WebHost:     host,
	}
}

// SaveDBPath persiste la ruta de base de datos en .vpn_config
func (c *Config) SaveDBPath(baseDir, newPath string) error {
	cfgFile := filepath.Join(baseDir, ".vpn_config")
	content := "DB_PATH=\"" + newPath + "\"\n"
	c.DBPath = newPath
	return os.WriteFile(cfgFile, []byte(content), 0644)
}
