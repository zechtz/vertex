// Package main
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

func main() {
	// Handle command line flags
	var showVersion bool
	var install bool
	var uninstall bool
	var port string
	var dataDir string
	var enableNginx bool
	var domain string
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&install, "install", false, "Install Vertex as a user service")
	flag.BoolVar(&uninstall, "uninstall", false, "Uninstall Vertex service")
	flag.BoolVar(&enableNginx, "nginx", false, "Configure nginx proxy for domain access (requires nginx to be installed)")
	flag.StringVar(&domain, "domain", "vertex.dev", "Domain name for nginx proxy (default: vertex.dev)")
	flag.StringVar(&port, "port", "8080", "Port to run the server on (default: 8080)")
	flag.StringVar(&dataDir, "data-dir", "", "Directory to store application data (database, logs, etc.). If not set, uses VERTEX_DATA_DIR environment variable or current directory")
	flag.Parse()

	if showVersion {
		fmt.Printf("Vertex %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	if install {
		if err := installService(enableNginx, domain); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}
		fmt.Println("✅ Vertex installed successfully as a user service!")
		fmt.Println("🚀 The service will start automatically.")
		if enableNginx {
			fmt.Printf("🌐 Access the web interface at: http://%s\n", domain)
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
func installService(enableNginx bool, domain string) error {
	installer := installer.NewServiceInstaller()
	if enableNginx {
		installer.SetDomain(domain)
		installer.EnableNginxProxy(true)
	}
	return installer.Install()
}

// uninstallService handles the --uninstall flag
func uninstallService() error {
	installer := installer.NewServiceInstaller()
	return installer.Uninstall()
}
