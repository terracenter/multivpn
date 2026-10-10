package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/terracenter/multivpn/internal/config"
	"github.com/terracenter/multivpn/internal/db"
	"github.com/terracenter/multivpn/internal/network"
	"github.com/terracenter/multivpn/internal/web"
)

var (
	Version = "1.0.0"
	GitHash = "dev"
)

func main() {
	fullVersion := fmt.Sprintf("%s-git-%s", Version, GitHash)

	// Obtener directorio del binario o de trabajo
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}

	cfg := config.LoadConfig(baseDir)

	// Inicializar SQLite y dump
	store, err := db.NewStorage(cfg.DBPath, cfg.DumpSQLPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] No se pudo inicializar la base de datos: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Inicializar Network Manager (Linux hoy, extensible a Windows/BSD)
	netMgr := network.NewLinuxManager()

	// Si no se pasaron argumentos o el primer argumento es "web", iniciar WebUI
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "web" {
		startWebUI(cfg, store, netMgr, fullVersion)
		return
	}

	command := args[0]
	switch command {
	case "status":
		runCLIStatus(store, netMgr, fullVersion)

	case "up":
		runCLIUp(args[1:], store, netMgr)

	case "down":
		runCLIDown(args[1:], store, netMgr)

	case "shield":
		runCLIShield(args[1:], store, netMgr)

	case "scan":
		runCLIScan(store, netMgr)

	case "version", "-v", "--version":
		fmt.Printf("MultiVPN v%s (GPLv3) - Plataforma: %s/%s\n", fullVersion, runtime.GOOS, runtime.GOARCH)

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Printf("Comando desconocido: '%s'. Ejecuta 'multivpn help' para ver las opciones disponibles.\n", command)
		os.Exit(1)
	}
}

func startWebUI(cfg *config.Config, store *db.Storage, netMgr network.VPNProvider, version string) {
	server, err := web.NewServer(cfg, store, netMgr, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Falla inicializando el servidor web: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("http://%s:%s", cfg.WebHost, cfg.WebPort)

	// Abrir navegador en una goroutine
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(url)
	}()

	fmt.Printf("🌐 MultiVPN v%s iniciado.\n", version)
	fmt.Printf("🔗 Panel de control disponible en: %s\n", url)
	fmt.Println("Presiona Ctrl+C para detener el servicio.")

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Servidor detenido: %v\n", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start()
}

func runCLIStatus(store *db.Storage, netMgr network.VPNProvider, version string) {
	fmt.Printf("=== MultiVPN v%s - Estado de Conexiones ===\n", version)
	conns, err := store.GetConnections()
	if err != nil {
		fmt.Printf("Error leyendo conexiones: %v\n", err)
		return
	}

	for _, c := range conns {
		active, _ := netMgr.IsConnectionActive(c.Name)
		hasRisk, _ := netMgr.HasDefaultGatewayRisk(c.Name)

		status := "⚪ Inactiva"
		if active {
			status = "🟢 Conectada"
		}

		shieldStatus := "🛡️ Split-Tunneling OK"
		if hasRisk {
			shieldStatus = "⚠️ RIESGO: Default Gateway Activo"
		}

		tipo := c.Type
		if c.IsSalida {
			tipo += " (Salida)"
		}

		fmt.Printf("• %-20s | %-12s | %-14s | %s\n", c.Name, tipo, status, shieldStatus)
	}
}

func runCLIUp(args []string, store *db.Storage, netMgr network.VPNProvider) {
	if len(args) == 0 {
		fmt.Println("Uso: multivpn up <nombre_vpn> | multivpn up --all")
		return
	}

	target := args[0]
	conns, _ := store.GetConnections()

	for _, c := range conns {
		if target == "--all" || strings.EqualFold(c.Name, target) {
			fmt.Printf("[*] Levantando VPN: %s...\n", c.Name)
			dbRoutes, _ := store.GetRoutes(c.ID, "")
			var routes []string
			for _, r := range dbRoutes {
				routes = append(routes, r.Route)
			}

			var resolvedIPs []string
			if c.IsSalida {
				dbDomains, _ := store.GetDomains(c.ID, "")
				var doms []string
				for _, d := range dbDomains {
					doms = append(doms, d.Domain)
				}
				resolvedIPs = network.ResolveDomainsParallel(doms, 3*time.Second)
			}

			if err := netMgr.Up(c, routes, resolvedIPs); err != nil {
				fmt.Printf("[ERROR] Falla conectando %s: %v\n", c.Name, err)
			} else {
				fmt.Printf("[OK] Conexión %s levantada exitosamente.\n", c.Name)
			}
		}
	}
}

func runCLIDown(args []string, store *db.Storage, netMgr network.VPNProvider) {
	target := "--all"
	if len(args) > 0 {
		target = args[0]
	}

	conns, _ := store.GetConnections()
	for _, c := range conns {
		if target == "--all" || strings.EqualFold(c.Name, target) {
			fmt.Printf("[*] Desconectando: %s...\n", c.Name)
			_ = netMgr.Down(c.Name)
		}
	}
	fmt.Println("[OK] Desconexión completada.")
}

func runCLIShield(args []string, store *db.Storage, netMgr network.VPNProvider) {
	target := "--all"
	if len(args) > 0 {
		target = args[0]
	}

	conns, _ := store.GetConnections()
	for _, c := range conns {
		if target == "--all" || strings.EqualFold(c.Name, target) {
			fmt.Printf("[*] Aplicando blindaje Split-Tunneling a: %s...\n", c.Name)
			if err := netMgr.ShieldConnection(c.Name); err != nil {
				fmt.Printf("[ERROR] No se pudo blindar %s: %v\n", c.Name, err)
			} else {
				fmt.Printf("[OK] %s blindada contra fuga de default gateway.\n", c.Name)
			}
		}
	}
}

func runCLIScan(store *db.Storage, netMgr network.VPNProvider) {
	fmt.Println("[*] Escaneando conexiones de NetworkManager...")
	detected, err := netMgr.ScanConnections()
	if err != nil {
		fmt.Printf("[ERROR] Error escaneando conexiones: %v\n", err)
		return
	}

	for _, d := range detected {
		isSalida := strings.Contains(strings.ToLower(d.Name), "salida") || d.Type == "wireguard"
		empresa := "general"
		if !isSalida {
			empresa = "trabajo"
		}
		_, _ = store.UpsertConnection(d.Name, d.Type, empresa, isSalida)
		_ = netMgr.ShieldConnection(d.Name)
		fmt.Printf("[OK] Detectada y blindada: %s (%s)\n", d.Name, d.Type)
	}
	fmt.Println("[OK] Escaneo y sincronización completados.")
}

func printHelp() {
	fmt.Println(`MultiVPN — Gestor de Conexiones VPN con WebUI Local (GPLv3)

USO:
  multivpn                    Inicia el servidor y abre la WebUI local en tu navegador
  multivpn web                Inicia la WebUI local explícitamente
  multivpn status             Muestra el estado de las conexiones en la terminal
  multivpn up <nombre>        Levanta una VPN específica con sus rutas
  multivpn up --all           Levanta todas las VPNs administradas
  multivpn down [nombre]      Desconecta una VPN o todas (--all)
  multivpn shield [nombre]    Aplica blindaje de Split-Tunneling a una o todas las VPNs
  multivpn scan               Escanea y sincroniza perfiles de NetworkManager con SQLite
  multivpn version            Muestra la versión actual y plataforma
  multivpn help               Muestra este mensaje de ayuda`)
}
