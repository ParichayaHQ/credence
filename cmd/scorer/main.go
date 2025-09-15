package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
	"github.com/ParichayaHQ/credence/internal/score"
	"github.com/ParichayaHQ/credence/internal/store"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	var (
		port     = flag.Int("port", 8082, "HTTP server port")
		dbPath   = flag.String("db", getEnvOrDefault("DATA_DIR", "./scorer-data"), "Database path")
	)
	flag.Parse()

	log.Printf("Starting Credence Trust Scorer on port %d", *port)

	// Initialize storage with simplified configuration
	storeConfig := store.DefaultConfig()
	storeConfig.BlobStore.FSPath = *dbPath + "/blobs"
	storeConfig.RocksDB.Path = *dbPath + "/rocksdb"

	blobStore, err := store.NewFilesystemBlobStore(&storeConfig.BlobStore)
	if err != nil {
		log.Fatalf("Failed to initialize blob store: %v", err)
	}
	defer blobStore.Close()

	// For now, we'll use the same blob store as event store
	// In production, you'd want separate storage or use RocksDB

	// Initialize data provider (mock implementation for now)
	dataProvider := NewMockDataProvider(blobStore)

	// Initialize cryptographic components
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		log.Fatalf("Failed to generate signing key: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)

	// Initialize scoring components
	config := score.DefaultScoreConfig()
	
	decayFunc := score.NewExponentialDecayFunction()
	budgetManager := score.NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := score.NewNetworkGraphAnalyzer(dataProvider)
	validator := NewMockValidator()

	// Initialize scoring engine
	engine := score.NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)

	// Initialize HTTP service
	httpService := score.NewHTTPService(
		engine,
		budgetManager,
		graphAnalyzer,
		config,
		*port,
	)

	// Start server in background
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on :%d", *port)
		if err := httpService.Start(); err != nil {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case <-interrupt:
		log.Println("Shutting down server...")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpService.Stop(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// MockDataProvider provides a mock implementation of score.DataProvider
type MockDataProvider struct {
	blobStore  store.BlobStore
	vouches    map[string][]*score.VouchData
	scores     map[string]*score.Score
}

func NewMockDataProvider(blobStore store.BlobStore) *MockDataProvider {
	return &MockDataProvider{
		blobStore:  blobStore,
		vouches:    make(map[string][]*score.VouchData),
		scores:     make(map[string]*score.Score),
	}
}

func (m *MockDataProvider) GetVouches(ctx context.Context, did, context string, maxEpoch int64) ([]*score.VouchData, error) {
	key := fmt.Sprintf("%s:%s", did, context)
	vouches, exists := m.vouches[key]
	if !exists {
		// Return some mock vouches for testing
		vouches = []*score.VouchData{
			{
				FromDID:   "did:key:example1",
				ToDID:     did,
				Context:   context,
				Strength:  10.0,
				Timestamp: time.Now().AddDate(0, 0, -30),
				Epoch:     maxEpoch - 30,
			},
			{
				FromDID:   "did:key:example2",
				ToDID:     did,
				Context:   context,
				Strength:  15.0,
				Timestamp: time.Now().AddDate(0, 0, -15),
				Epoch:     maxEpoch - 15,
			},
		}
		m.vouches[key] = vouches
	}
	
	// Filter by epoch
	filtered := make([]*score.VouchData, 0)
	for _, vouch := range vouches {
		if vouch.Epoch <= maxEpoch {
			filtered = append(filtered, vouch)
		}
	}
	
	return filtered, nil
}

func (m *MockDataProvider) GetAttestations(ctx context.Context, did, context string, maxEpoch int64) ([]*score.AttestationData, error) {
	// Return mock attestations
	return []*score.AttestationData{
		{
			DID:              did,
			Context:          context,
			Type:             "employment",
			IssuerDID:        "did:key:employer123",
			IssuerReputation: 0.8,
			Weight:           20.0,
			Timestamp:        time.Now().AddDate(0, -6, 0),
			Epoch:            maxEpoch - 180,
		},
	}, nil
}

func (m *MockDataProvider) GetReports(ctx context.Context, did, context string, maxEpoch int64) ([]*score.ReportData, error) {
	// Return mock reports (none by default for clean scores)
	return []*score.ReportData{}, nil
}

func (m *MockDataProvider) GetKYCData(ctx context.Context, did, context string, maxEpoch int64) ([]*score.KYCData, error) {
	// Return mock KYC data
	return []*score.KYCData{
		{
			DID:       did,
			Context:   context,
			Type:      "kyc_level_2",
			Level:     2,
			IssuerDID: "did:key:kyc_provider",
			Weight:    25.0,
			Timestamp: time.Now().AddDate(0, -12, 0),
			Epoch:     maxEpoch - 365,
		},
	}, nil
}

func (m *MockDataProvider) GetTimeData(ctx context.Context, did, context string, maxEpoch int64) (*score.TimeData, error) {
	// Return mock time data
	return &score.TimeData{
		DID:           did,
		Context:       context,
		FirstActivity: time.Now().AddDate(-2, 0, 0), // 2 years ago
		LastActivity:  time.Now().AddDate(0, 0, -7), // 1 week ago
		ActivityCount: 156,
		Epoch:         maxEpoch,
	}, nil
}

func (m *MockDataProvider) GetScore(ctx context.Context, did, context string, epoch int64) (*score.Score, error) {
	key := fmt.Sprintf("%s:%s:%d", did, context, epoch)
	if score, exists := m.scores[key]; exists {
		return score, nil
	}
	return nil, fmt.Errorf("score not found")
}

func (m *MockDataProvider) StoreScore(ctx context.Context, score *score.Score) error {
	key := fmt.Sprintf("%s:%s:%d", score.DID, score.Context, score.Epoch)
	m.scores[key] = score
	return nil
}

// MockValidator provides a mock implementation of score.ScoreValidator
type MockValidator struct{}

func NewMockValidator() score.ScoreValidator {
	return &MockValidator{}
}

func (v *MockValidator) ValidateInputData(ctx context.Context, did, context string, epoch int64) error {
	if did == "" {
		return fmt.Errorf("DID cannot be empty")
	}
	if epoch < 0 {
		return fmt.Errorf("epoch cannot be negative")
	}
	return nil
}

func (v *MockValidator) ValidateScoreRange(score float64) error {
	if score < 0 {
		return fmt.Errorf("score cannot be negative")
	}
	if score > 1000 {
		return fmt.Errorf("score suspiciously high: %.2f", score)
	}
	return nil
}

func (v *MockValidator) ValidateComponents(components *score.ScoreComponents, factors *score.ScoreFactors) error {
	if components.K < 0 || components.A < 0 || components.V < 0 || components.R < 0 || components.T < 0 {
		return fmt.Errorf("component values cannot be negative")
	}
	return nil
}

func (v *MockValidator) ValidateIntegrity(ctx context.Context, score *score.Score) error {
	// Basic integrity checks
	if score.DID == "" {
		return fmt.Errorf("score DID cannot be empty")
	}
	if score.Value < 0 {
		return fmt.Errorf("score value cannot be negative")
	}
	if score.Epoch < 0 {
		return fmt.Errorf("score epoch cannot be negative")
	}
	return nil
}