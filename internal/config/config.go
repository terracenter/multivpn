package config

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config almacena las configuraciones globales de MultiVPN
type Config struct {
	ConfigDir   string
	DBPath      string
	DumpSQLPath string
	WebPort     string
	WebHost     string
}

// LoadConfig obtiene la configuración respetando el estándar XDG en Linux/BSD y AppData en Windows
func LoadConfig(baseDir string) *Config {
	configDir := getDefaultConfigDir()
	defaultDB := getDefaultDBPath(configDir)
	dumpFile := filepath.Join(baseDir, "data", "vpn_data_dump.sql")

	// Si no existe el dump en baseDir, buscar relativo al binario
	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			exeDump := filepath.Join(filepath.Dir(exe), "data", "vpn_data_dump.sql")
			if _, err := os.Stat(exeDump); err == nil {
				dumpFile = exeDump
			}
		}
	}

	dbPath := defaultDB

	// Leer config.env si existe en el directorio de configuración
	cfgFile := filepath.Join(configDir, "config.env")
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

	// Permitir override por variable de entorno
	if envDB := os.Getenv("MULTIVPN_DB_PATH"); envDB != "" {
		dbPath = envDB
	}

	// Asegurar existencia del directorio de base de datos
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	port := os.Getenv("MULTIVPN_PORT")
	if port == "" {
		port = "8765"
	}

	host := os.Getenv("MULTIVPN_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	return &Config{
		ConfigDir:   configDir,
		DBPath:      dbPath,
		DumpSQLPath: dumpFile,
		WebPort:     port,
		WebHost:     host,
	}
}

func getDefaultConfigDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "multivpn")
	}

	// Linux, BSD, Darwin (XDG Base Directory)
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		home, _ := os.UserHomeDir()
		xdgConfig = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfig, "multivpn")
}

func getDefaultDBPath(configDir string) string {
	return filepath.Join(configDir, "vpn_data.db")
}

// SaveDBPath persiste una ruta personalizada en config.env
func (c *Config) SaveDBPath(newPath string) error {
	_ = os.MkdirAll(c.ConfigDir, 0755)
	cfgFile := filepath.Join(c.ConfigDir, "config.env")
	content := "DB_PATH=\"" + newPath + "\"\n"
	c.DBPath = newPath
	return os.WriteFile(cfgFile, []byte(content), 0644)
}
