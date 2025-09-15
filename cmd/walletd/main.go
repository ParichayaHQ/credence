package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ParichayaHQ/credence/cmd/walletd/server"
	"github.com/ParichayaHQ/credence/internal/wallet"
)

var (
	port     = flag.String("port", "8080", "HTTP server port")
	host     = flag.String("host", "127.0.0.1", "HTTP server host")
	logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	dataDir  = flag.String("data-dir", "", "Data directory for wallet storage (defaults to OS-specific location)")
)

func main() {
	flag.Parse()

	// Set up logging
	setupLogging(*logLevel)

	// Initialize wallet service
	walletService, err := initializeWallet(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize wallet: %v", err)
	}

	// Create HTTP server
	srv := server.NewServer(walletService)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", *host, *port),
		Handler:      srv.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting walletd HTTP server on %s:%s", *host, *port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down walletd server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Close wallet service
	if err := walletService.Close(); err != nil {
		log.Printf("Error closing wallet service: %v", err)
	}

	log.Println("Walletd server stopped")
}

func initializeWallet(dataDir string) (*wallet.Service, error) {
	// Use default data directory if not specified
	if dataDir == "" {
		var err error
		dataDir, err = getDefaultDataDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get default data directory: %w", err)
		}
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	log.Printf("Using data directory: %s", dataDir)

	// Initialize wallet service
	config := &wallet.Config{
		DataDir: dataDir,
		// Add other configuration options as needed
	}

	walletService, err := wallet.NewService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet service: %w", err)
	}

	return walletService, nil
}

func getDefaultDataDir() (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Platform-specific data directory
	var dataDir string
	switch {
	case isWindows():
		dataDir = fmt.Sprintf("%s\\AppData\\Roaming\\Credence\\wallet", homeDir)
	case isMacOS():
		dataDir = fmt.Sprintf("%s/Library/Application Support/Credence/wallet", homeDir)
	default: // Linux and other Unix-like systems
		dataDir = fmt.Sprintf("%s/.local/share/credence/wallet", homeDir)
	}

	return dataDir, nil
}

func isWindows() bool {
	return os.Getenv("OS") == "Windows_NT"
}

func isMacOS() bool {
	return os.Getenv("HOME") != "" && fileExists("/System/Library/CoreServices/SystemVersion.plist")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func setupLogging(level string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	switch level {
	case "debug":
		log.SetOutput(os.Stdout)
	case "info":
		log.SetOutput(os.Stdout)
	case "warn", "error":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stdout)
	}

	log.Printf("Log level set to: %s", level)
}