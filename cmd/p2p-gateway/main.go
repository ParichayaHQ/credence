package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ParichayaHQ/credence/internal/p2p"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	var (
		listenAddr   = flag.String("listen", "/ip4/0.0.0.0/tcp/4001", "P2P listen address")
		httpAddr     = flag.String("http", ":8080", "HTTP bridge listen address")
		bootstrap    = flag.String("bootstrap", "", "Bootstrap peer addresses (comma-separated)")
		dhtMode      = flag.String("dht-mode", "auto", "DHT mode: client, server, auto")
	)
	flag.Parse()

	// Create logger for main application
	logger := p2p.NewLogger("P2PGateway", p2p.LogLevelInfo)
	
	logger.Info("Starting P2P Gateway", map[string]interface{}{
		"listen_addr": *listenAddr,
		"http_addr":   *httpAddr,
		"dht_mode":    *dhtMode,
		"bootstrap":   *bootstrap,
	})

	// Parse listen address
	listenMA, err := multiaddr.NewMultiaddr(*listenAddr)
	if err != nil {
		logger.Fatal("Invalid listen address", map[string]interface{}{
			"listen_addr": *listenAddr,
			"error": err,
		})
	}

	// Create P2P configuration
	config := p2p.DefaultConfig()
	config.ListenAddrs = []multiaddr.Multiaddr{listenMA}
	config.DHTConfig.Mode = *dhtMode

	// Parse bootstrap peers if provided
	if *bootstrap != "" {
		// Would parse comma-separated bootstrap addresses
		logger.Info("Bootstrap peers configured", map[string]interface{}{
			"peers": *bootstrap,
			"note": "parsing not yet implemented",
		})
	}

	// Create P2P host
	p2pHost := p2p.NewP2PHost(config)

	// Create HTTP bridge
	bridge := p2p.NewHTTPBridge(p2pHost, *httpAddr)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start P2P host
	logger.Info("Starting P2P host", map[string]interface{}{"addr": listenMA.String()})
	if err := p2pHost.Start(ctx); err != nil {
		logger.Fatal("Failed to start P2P host", map[string]interface{}{"error": err})
	}
	defer func() {
		if err := p2pHost.Stop(ctx); err != nil {
			logger.Error("Error stopping P2P host during cleanup", map[string]interface{}{"error": err})
		}
	}()

	// Start HTTP bridge
	logger.Info("Starting HTTP bridge", map[string]interface{}{"addr": *httpAddr})
	if err := bridge.Start(ctx); err != nil {
		logger.Fatal("Failed to start HTTP bridge", map[string]interface{}{"error": err})
	}
	defer func() {
		if err := bridge.Stop(ctx); err != nil {
			logger.Error("Error stopping HTTP bridge during cleanup", map[string]interface{}{"error": err})
		}
	}()

	// Print network information
	netInfo := p2pHost.GetNetworkInfo()
	logger.Info("P2P Gateway started successfully", map[string]interface{}{
		"peer_id":         netInfo["peer_id"],
		"connected_peers": netInfo["connected_peers"],
		"http_addr":       *httpAddr,
		"status":          netInfo["status"],
	})

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received shutdown signal", map[string]interface{}{"signal": sig.String()})

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop services
	logger.Info("Stopping services gracefully")
	
	if err := bridge.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping HTTP bridge", map[string]interface{}{"error": err})
	}

	if err := p2pHost.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping P2P host", map[string]interface{}{"error": err})
	}

	logger.Info("P2P Gateway stopped gracefully")
}