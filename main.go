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
	"github.com/zechtz/nest-up/internal/config"
	"github.com/zechtz/nest-up/internal/database"
	"github.com/zechtz/nest-up/internal/handlers"
	"github.com/zechtz/nest-up/internal/services"
	"github.com/zechtz/nest-up/web"
)

// Version information (set by build flags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Handle version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("NeST Service Manager %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Display startup information
	logMessage(fmt.Sprintf("Starting NeST Service Manager %s", version))

	// Initialize database first
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

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
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		logMessage("Starting NeST Service Manager on :8080")
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