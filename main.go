package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/config"
	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/handlers"
	"github.com/zechtz/vertex/internal/installer"
	"github.com/zechtz/vertex/internal/services"
	"github.com/zechtz/vertex/web"
)

// Version information (set by build flags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// parseSubcommands converts subcommands to equivalent flags for backward compatibility
// Supports both "vertex start" and "vertex --start" syntax
func parseSubcommands() {
	if len(os.Args) < 2 {
		return
	}

	subcommand := os.Args[1]
	
	// Skip if first arg is already a flag (starts with -)
	if strings.HasPrefix(subcommand, "-") {
		return
	}

	// Map subcommands to their equivalent flags
	subcommandMap := map[string]string{
		"start":     "--start",
		"stop":      "--stop", 
		"restart":   "--restart",
		"status":    "--status",
		"logs":      "--logs",
		"install":   "--install",
		"uninstall": "--uninstall",
		"update":    "--update",
		"version":   "--version",
	}

	// Check if the subcommand is valid
	if flag, exists := subcommandMap[subcommand]; exists {
		// Replace the subcommand with the equivalent flag
		os.Args[1] = flag
		
		// Handle special case for 'logs' subcommand with -f or --follow
		if subcommand == "logs" && len(os.Args) > 2 {
			for i := 2; i < len(os.Args); i++ {
				if os.Args[i] == "-f" {
					os.Args[i] = "--follow"
				}
			}
		}
	}
}

func main() {
	// Parse subcommands before flag parsing
	parseSubcommands()
	
	// Handle command line flags
	var showVersion bool
	var install bool
	var uninstall bool
	var update bool
	var start bool
	var stop bool
	var restart bool
	var status bool
	var logs bool
	var follow bool
	var port string
	var dataDir string
	var enableNginx bool
	var enableHTTPS bool
	var domain string
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&install, "install", false, "Install Vertex as a user service")
	flag.BoolVar(&uninstall, "uninstall", false, "Uninstall Vertex service")
	flag.BoolVar(&update, "update", false, "Update the Vertex service")
	flag.BoolVar(&start, "start", false, "Start the Vertex service")
	flag.BoolVar(&stop, "stop", false, "Stop the Vertex service")
	flag.BoolVar(&restart, "restart", false, "Restart the Vertex service")
	flag.BoolVar(&status, "status", false, "Show service status")
	flag.BoolVar(&logs, "logs", false, "Show service logs")
	flag.BoolVar(&follow, "follow", false, "Follow log output (use with --logs)")
	flag.BoolVar(&enableNginx, "nginx", false, "Configure nginx proxy for domain access (requires nginx to be installed)")
	flag.BoolVar(&enableHTTPS, "https", false, "Enable HTTPS with locally-trusted certificates (automatically enabled for .dev domains)")
	flag.StringVar(&domain, "domain", "vertex.dev", "Domain name for nginx proxy (automatically installs with nginx when specified)")
	flag.StringVar(&port, "port", "54321", "Port to run the server on (default: 54321)")
	flag.StringVar(&dataDir, "data-dir", "", "Directory to store application data (database, logs, etc.). If not set, uses VERTEX_DATA_DIR environment variable or current directory")
	
	// Custom usage function to show both flag and subcommand syntax
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSubcommands (recommended):\n")
		fmt.Fprintf(os.Stderr, "  vertex start        Start the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  vertex stop         Stop the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  vertex restart      Restart the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  vertex status       Show service status\n")
		fmt.Fprintf(os.Stderr, "  vertex logs         Show service logs\n")
		fmt.Fprintf(os.Stderr, "  vertex logs -f      Follow log output (tail -f style)\n")
		fmt.Fprintf(os.Stderr, "  vertex install      Install Vertex as a user service\n")
		fmt.Fprintf(os.Stderr, "  vertex uninstall    Uninstall Vertex service\n")
		fmt.Fprintf(os.Stderr, "  vertex update       Update the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  vertex version      Show version information\n")
		fmt.Fprintf(os.Stderr, "\nFlags (alternative syntax):\n")
		fmt.Fprintf(os.Stderr, "  --data-dir string\n")
		fmt.Fprintf(os.Stderr, "    \tDirectory to store application data (database, logs, etc.). If not set, uses VERTEX_DATA_DIR environment variable or current directory\n")
		fmt.Fprintf(os.Stderr, "  --domain string\n")
		fmt.Fprintf(os.Stderr, "    \tDomain name for nginx proxy (automatically installs with nginx when specified) (default \"vertex.dev\")\n")
		fmt.Fprintf(os.Stderr, "  --follow\n")
		fmt.Fprintf(os.Stderr, "    \tFollow log output (use with --logs)\n")
		fmt.Fprintf(os.Stderr, "  --https\n")
		fmt.Fprintf(os.Stderr, "    \tEnable HTTPS with locally-trusted certificates (automatically enabled for .dev domains)\n")
		fmt.Fprintf(os.Stderr, "  --install\n")
		fmt.Fprintf(os.Stderr, "    \tInstall Vertex as a user service\n")
		fmt.Fprintf(os.Stderr, "  --logs\n")
		fmt.Fprintf(os.Stderr, "    \tShow service logs\n")
		fmt.Fprintf(os.Stderr, "  --nginx\n")
		fmt.Fprintf(os.Stderr, "    \tConfigure nginx proxy for domain access (requires nginx to be installed)\n")
		fmt.Fprintf(os.Stderr, "  --port string\n")
		fmt.Fprintf(os.Stderr, "    \tPort to run the server on (default: 54321) (default \"54321\")\n")
		fmt.Fprintf(os.Stderr, "  --restart\n")
		fmt.Fprintf(os.Stderr, "    \tRestart the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  --start\n")
		fmt.Fprintf(os.Stderr, "    \tStart the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  --status\n")
		fmt.Fprintf(os.Stderr, "    \tShow service status\n")
		fmt.Fprintf(os.Stderr, "  --stop\n")
		fmt.Fprintf(os.Stderr, "    \tStop the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  --uninstall\n")
		fmt.Fprintf(os.Stderr, "    \tUninstall Vertex service\n")
		fmt.Fprintf(os.Stderr, "  --update\n")
		fmt.Fprintf(os.Stderr, "    \tUpdate the Vertex service\n")
		fmt.Fprintf(os.Stderr, "  --version\n")
		fmt.Fprintf(os.Stderr, "    \tShow version information\n")
	}
	
	flag.Parse()

	if showVersion {
		fmt.Printf("Vertex %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	if update {
		if err := installer.UpdateService(); err != nil {
			log.Fatalf("Failed to update service: %v", err)
		}
		os.Exit(0)
	}

	if start {
		if err := startService(); err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
		fmt.Println("✅ Vertex service started successfully!")
		os.Exit(0)
	}

	if stop {
		if err := stopService(); err != nil {
			log.Fatalf("Failed to stop service: %v", err)
		}
		fmt.Println("✅ Vertex service stopped successfully!")
		os.Exit(0)
	}

	if restart {
		if err := restartService(); err != nil {
			log.Fatalf("Failed to restart service: %v", err)
		}
		fmt.Println("✅ Vertex service restarted successfully!")
		os.Exit(0)
	}

	if status {
		if err := showStatus(); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}
		os.Exit(0)
	}

	if logs {
		if err := showLogs(follow); err != nil {
			log.Fatalf("Failed to show logs: %v", err)
		}
		os.Exit(0)
	}

	// Check if domain flag was explicitly specified (smart auto-install)
	domainWasExplicitlySet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "domain" {
			domainWasExplicitlySet = true
		}
	})
	
	// Auto-enable HTTPS for .dev domains (Google-owned TLD requires HTTPS)
	if strings.HasSuffix(domain, ".dev") && !enableHTTPS {
		enableHTTPS = true
		fmt.Printf("🔒 .dev domain detected (%s), automatically enabling HTTPS\n", domain)
	}
	
	// Auto-install with nginx if domain is specified
	if domainWasExplicitlySet && !install && !uninstall {
		install = true
		enableNginx = true
		fmt.Printf("🌐 Domain specified (%s), automatically installing with nginx proxy\n", domain)
	}
	
	// Auto-enable nginx if HTTPS is requested
	if enableHTTPS && !enableNginx {
		enableNginx = true
		fmt.Printf("🔒 HTTPS enabled, automatically configuring nginx proxy\n")
	}

	if install {
		// Auto-enable nginx if domain flag was explicitly specified (smart UX)
		if domainWasExplicitlySet && !enableNginx {
			enableNginx = true
			fmt.Printf("🌐 Domain specified (%s), automatically enabling nginx proxy\n", domain)
		}
		
		if err := installService(enableNginx, enableHTTPS, domain); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}
		fmt.Println("✅ Vertex installed successfully as a user service!")
		fmt.Println("🚀 The service will start automatically.")
		if enableNginx {
			protocol := "http"
			if enableHTTPS {
				protocol = "https"
			}
			fmt.Printf("🌐 Access the web interface at: %s://%s\n", protocol, domain)
			fmt.Printf("   Also available at: http://localhost:%s\n", port)
		} else {
			fmt.Printf("🌐 Access the web interface at: http://localhost:%s\n", port)
			fmt.Println("   💡 Use --nginx flag next time to configure domain access")
		}
		os.Exit(0)
	}

	if uninstall {
		if err := uninstallService(); err != nil {
			log.Fatalf("Uninstallation failed: %v", err)
		}
		fmt.Println("✅ Vertex service uninstalled successfully!")
		fmt.Println("🗑️ All service files and data have been removed.")
		os.Exit(0)
	}

	// Set data directory if provided
	if dataDir != "" {
		os.Setenv("VERTEX_DATA_DIR", dataDir)
	}

	// Display startup information
	logMessage(fmt.Sprintf("Starting Vertex %s", version))
	if dataDir := os.Getenv("VERTEX_DATA_DIR"); dataDir != "" {
		logMessage(fmt.Sprintf("Using data directory: %s", dataDir))
	}

	// Initialize database first
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Detect and setup Java environment
	javaEnv := services.DetectJavaEnvironment()
	if javaEnv.Available {
		if err := javaEnv.SetupJavaEnvironment(); err != nil {
			log.Printf("[WARN] Failed to setup Java environment: %v", err)
		}
	} else {
		log.Printf("[WARN] Java not detected: %s", javaEnv.ErrorMsg)
		log.Printf("[INFO] Services requiring Java may fail to start")
	}

	// Automatic environment setup check with database
	checkAndSetupEnvironment(db)

	// Load environment variables
	if err := config.LoadEnvironmentVariables(); err != nil {
		log.Printf("Warning: Could not load environment variables: %v", err)
	}

	// Load configuration
	cfg := config.LoadDefaultConfig()

	// Initialize service manager
	sm, err := services.NewManager(cfg, db)
	if err != nil {
		log.Fatal("Failed to create service manager:", err)
	}

	// Initialize handlers
	handler := handlers.NewHandler(sm)

	// Setup routes
	r := mux.NewRouter()
	handler.RegisterRoutes(r)

	// Serve embedded frontend assets
	uiFS, err := fs.Sub(web.EmbeddedUI, "dist")
	if err != nil {
		log.Fatal("Failed to access embedded UI:", err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(uiFS)))

	// Create HTTP server
	serverAddr := ":" + port
	server := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		logMessage(fmt.Sprintf("Starting Vertex on %s", serverAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	logMessage("Shutdown signal received, stopping all services...")

	// Stop all running services
	sm.GracefulShutdown()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		logMessage("Server shutdown complete")
	}
}

func logMessage(message string) {
	fmt.Printf("[INFO] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func checkAndSetupEnvironment(db *database.Database) {
	logMessage("Checking environment setup...")

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return
	}

	// Create environment setup instance with database
	envSetup := services.NewEnvironmentSetup(workingDir, db)

	// Initialize default environment variables in database
	logMessage("Initializing default environment variables in database...")
	if err := envSetup.InitializeDefaultEnvironmentVariables(); err != nil {
		log.Printf("Warning: Failed to initialize database environment variables: %v", err)
	}

	// Check environment status
	status := envSetup.CheckEnvironmentStatus()
	missingCount := status["missing"].(int)
	totalCount := status["total"].(int)

	if missingCount > 0 {
		logMessage(fmt.Sprintf("Environment setup needed: %d/%d variables missing", missingCount, totalCount))

		// Setup environment variables (now loads from database)
		logMessage("Setting up environment variables...")
		result := envSetup.SetupEnvironment()
		if result.Success {
			logMessage(fmt.Sprintf("Successfully set up %d environment variables", result.VariablesSet))
		} else {
			log.Printf("Warning: Environment setup failed: %s", result.Message)
		}

		if len(result.Errors) > 0 {
			log.Printf("Environment setup warnings: %v", result.Errors)
		}
	} else {
		logMessage("Environment already configured")
	}
}

// installService handles the --install flag
func installService(enableNginx bool, enableHTTPS bool, domain string) error {
	installer := installer.NewServiceInstaller()
	if enableNginx {
		installer.SetDomain(domain)
		installer.EnableNginxProxy(true)
		installer.EnableHTTPS(enableHTTPS)
	}
	return installer.Install()
}

// uninstallService handles the --uninstall flag
func uninstallService() error {
	installer := installer.NewServiceInstaller()
	return installer.Uninstall()
}

// startService handles the --start flag
func startService() error {
	serviceManager := installer.NewServiceManager()
	return serviceManager.Start()
}

// stopService handles the --stop flag
func stopService() error {
	serviceManager := installer.NewServiceManager()
	return serviceManager.Stop()
}

// restartService handles the --restart flag
func restartService() error {
	serviceManager := installer.NewServiceManager()
	return serviceManager.Restart()
}

// showStatus handles the --status flag
func showStatus() error {
	serviceManager := installer.NewServiceManager()
	return serviceManager.ShowStatus()
}

// showLogs handles the --logs flag
func showLogs(follow bool) error {
	serviceManager := installer.NewServiceManager()
	return serviceManager.ShowLogs(follow)
}
