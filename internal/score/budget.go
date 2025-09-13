package score

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// MemoryBudgetManager implements BudgetManager using in-memory storage
// In production, this would be backed by persistent storage
type MemoryBudgetManager struct {
	budgets    map[string]*VouchBudget // key: did:context:epoch
	mu         sync.RWMutex
	config     *ScoreConfig
	dataProvider DataProvider
}

// NewMemoryBudgetManager creates a new memory-based budget manager
func NewMemoryBudgetManager(config *ScoreConfig, dataProvider DataProvider) BudgetManager {
	return &MemoryBudgetManager{
		budgets:      make(map[string]*VouchBudget),
		config:       config,
		dataProvider: dataProvider,
	}
}

// getBudgetKey creates a unique key for budget storage
func (m *MemoryBudgetManager) getBudgetKey(did, context string, epoch int64) string {
	return fmt.Sprintf("%s:%s:%d", did, context, epoch)
}

// GetBudget implements BudgetManager.GetBudget
func (m *MemoryBudgetManager) GetBudget(ctx context.Context, did, context string, epoch int64) (*VouchBudget, error) {
	m.mu.RLock()
	key := m.getBudgetKey(did, context, epoch)
	if budget, exists := m.budgets[key]; exists {
		// Return a copy to prevent external modification
		budgetCopy := *budget
		m.mu.RUnlock()
		return &budgetCopy, nil
	}
	m.mu.RUnlock()
	
	// Budget doesn't exist, create it based on the DID's score
	return m.createBudget(ctx, did, context, epoch)
}

// createBudget creates a new budget for a DID based on their current score
func (m *MemoryBudgetManager) createBudget(ctx context.Context, did, context string, epoch int64) (*VouchBudget, error) {
	// Get the DID's current score to determine budget
	var score float64
	if cachedScore, err := m.dataProvider.GetScore(ctx, did, context, epoch); err == nil && cachedScore != nil {
		score = cachedScore.Value
	} else {
		// If no score available, use a default
		score = 10.0
	}
	
	// Calculate total budget: base + (score * multiplier)
	totalBudget := m.config.BaseBudget + (score * m.config.BudgetMultiplier)
	
	// Apply bounds
	if totalBudget < m.config.BaseBudget {
		totalBudget = m.config.BaseBudget
	}
	
	// Calculate reputation bond (portion of score at risk)
	reputationBond := score * 0.1 // 10% of score as bond
	if reputationBond > totalBudget * 0.5 {
		reputationBond = totalBudget * 0.5 // Maximum 50% of budget as bond
	}
	
	budget := &VouchBudget{
		DID:             did,
		Context:         context,
		Epoch:           epoch,
		TotalBudget:     totalBudget,
		SpentBudget:     0.0,
		RemainingBudget: totalBudget,
		ReputationBond:  reputationBond,
	}
	
	// Store the new budget
	m.mu.Lock()
	key := m.getBudgetKey(did, context, epoch)
	m.budgets[key] = budget
	m.mu.Unlock()
	
	// Return a copy
	budgetCopy := *budget
	return &budgetCopy, nil
}

// SpendBudget implements BudgetManager.SpendBudget
func (m *MemoryBudgetManager) SpendBudget(ctx context.Context, did, context string, epoch int64, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("spend amount must be positive")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.getBudgetKey(did, context, epoch)
	budget, exists := m.budgets[key]
	
	if !exists {
		// Create budget first
		m.mu.Unlock()
		newBudget, err := m.createBudget(ctx, did, context, epoch)
		if err != nil {
			return fmt.Errorf("failed to create budget: %w", err)
		}
		m.mu.Lock()
		budget = newBudget
		m.budgets[key] = budget
	}
	
	// Check if there's enough budget remaining
	if budget.RemainingBudget < amount {
		return fmt.Errorf("insufficient budget: available=%.2f, requested=%.2f", 
			budget.RemainingBudget, amount)
	}
	
	// Spend the budget
	budget.SpentBudget += amount
	budget.RemainingBudget -= amount
	
	return nil
}

// RefillBudget implements BudgetManager.RefillBudget
func (m *MemoryBudgetManager) RefillBudget(ctx context.Context, did, context string, epoch int64, score float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.getBudgetKey(did, context, epoch)
	budget, exists := m.budgets[key]
	
	if !exists {
		// If budget doesn't exist, create it
		m.mu.Unlock()
		_, err := m.createBudget(ctx, did, context, epoch)
		if err != nil {
			return fmt.Errorf("failed to create budget during refill: %w", err)
		}
		return nil
	}
	
	// Recalculate total budget based on new score
	newTotalBudget := m.config.BaseBudget + (score * m.config.BudgetMultiplier)
	
	// Apply bounds
	if newTotalBudget < m.config.BaseBudget {
		newTotalBudget = m.config.BaseBudget
	}
	
	// Update budget if the new total is higher
	if newTotalBudget > budget.TotalBudget {
		additionalBudget := newTotalBudget - budget.TotalBudget
		budget.TotalBudget = newTotalBudget
		budget.RemainingBudget += additionalBudget
	}
	
	// Update reputation bond
	budget.ReputationBond = score * 0.1
	if budget.ReputationBond > budget.TotalBudget * 0.5 {
		budget.ReputationBond = budget.TotalBudget * 0.5
	}
	
	return nil
}

// GetSpentBudget implements BudgetManager.GetSpentBudget
func (m *MemoryBudgetManager) GetSpentBudget(ctx context.Context, did, context string, epoch int64) (float64, error) {
	budget, err := m.GetBudget(ctx, did, context, epoch)
	if err != nil {
		return 0, err
	}
	
	return budget.SpentBudget, nil
}

// ValidateBudget implements BudgetManager.ValidateBudget
func (m *MemoryBudgetManager) ValidateBudget(ctx context.Context, did, context string, epoch int64) error {
	budget, err := m.GetBudget(ctx, did, context, epoch)
	if err != nil {
		return fmt.Errorf("failed to get budget: %w", err)
	}
	
	// Validate budget consistency
	if budget.SpentBudget < 0 {
		return fmt.Errorf("spent budget cannot be negative: %.2f", budget.SpentBudget)
	}
	
	if budget.RemainingBudget < 0 {
		return fmt.Errorf("remaining budget cannot be negative: %.2f", budget.RemainingBudget)
	}
	
	if budget.TotalBudget < 0 {
		return fmt.Errorf("total budget cannot be negative: %.2f", budget.TotalBudget)
	}
	
	// Check if spent + remaining equals total (within small tolerance for floating point)
	calculatedTotal := budget.SpentBudget + budget.RemainingBudget
	tolerance := 0.001
	if math.Abs(calculatedTotal - budget.TotalBudget) > tolerance {
		return fmt.Errorf("budget inconsistency: spent(%.2f) + remaining(%.2f) != total(%.2f)", 
			budget.SpentBudget, budget.RemainingBudget, budget.TotalBudget)
	}
	
	// Validate reputation bond
	if budget.ReputationBond < 0 {
		return fmt.Errorf("reputation bond cannot be negative: %.2f", budget.ReputationBond)
	}
	
	return nil
}

// PersistentBudgetManager implements BudgetManager with persistent storage
type PersistentBudgetManager struct {
	memoryManager *MemoryBudgetManager
	storage       BudgetStorage
	config        *ScoreConfig
}

// BudgetStorage defines the interface for budget persistence
type BudgetStorage interface {
	StoreBudget(ctx context.Context, budget *VouchBudget) error
	GetBudget(ctx context.Context, did, context string, epoch int64) (*VouchBudget, error)
	DeleteBudget(ctx context.Context, did, context string, epoch int64) error
	ListBudgets(ctx context.Context, did, context string, fromEpoch, toEpoch int64) ([]*VouchBudget, error)
}

// NewPersistentBudgetManager creates a budget manager with persistent storage
func NewPersistentBudgetManager(config *ScoreConfig, dataProvider DataProvider, storage BudgetStorage) BudgetManager {
	return &PersistentBudgetManager{
		memoryManager: NewMemoryBudgetManager(config, dataProvider).(*MemoryBudgetManager),
		storage:       storage,
		config:        config,
	}
}

// GetBudget implements BudgetManager.GetBudget with persistence
func (p *PersistentBudgetManager) GetBudget(ctx context.Context, did, context string, epoch int64) (*VouchBudget, error) {
	// Try memory first for performance
	if budget, err := p.memoryManager.GetBudget(ctx, did, context, epoch); err == nil {
		return budget, nil
	}
	
	// Try persistent storage
	budget, err := p.storage.GetBudget(ctx, did, context, epoch)
	if err == nil {
		// Cache in memory for future access
		key := p.memoryManager.getBudgetKey(did, context, epoch)
		p.memoryManager.mu.Lock()
		p.memoryManager.budgets[key] = budget
		p.memoryManager.mu.Unlock()
		
		budgetCopy := *budget
		return &budgetCopy, nil
	}
	
	// Create new budget
	return p.memoryManager.createBudget(ctx, did, context, epoch)
}

// SpendBudget implements BudgetManager.SpendBudget with persistence
func (p *PersistentBudgetManager) SpendBudget(ctx context.Context, did, context string, epoch int64, amount float64) error {
	// Spend in memory first
	if err := p.memoryManager.SpendBudget(ctx, did, context, epoch, amount); err != nil {
		return err
	}
	
	// Get updated budget and persist
	budget, err := p.memoryManager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		return fmt.Errorf("failed to get updated budget: %w", err)
	}
	
	// Persist to storage
	if err := p.storage.StoreBudget(ctx, budget); err != nil {
		// Try to rollback memory change
		p.memoryManager.mu.Lock()
		key := p.memoryManager.getBudgetKey(did, context, epoch)
		if memBudget, exists := p.memoryManager.budgets[key]; exists {
			memBudget.SpentBudget -= amount
			memBudget.RemainingBudget += amount
		}
		p.memoryManager.mu.Unlock()
		
		return fmt.Errorf("failed to persist budget: %w", err)
	}
	
	return nil
}

// RefillBudget implements BudgetManager.RefillBudget with persistence
func (p *PersistentBudgetManager) RefillBudget(ctx context.Context, did, context string, epoch int64, score float64) error {
	if err := p.memoryManager.RefillBudget(ctx, did, context, epoch, score); err != nil {
		return err
	}
	
	// Persist updated budget
	budget, err := p.memoryManager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		return fmt.Errorf("failed to get updated budget: %w", err)
	}
	
	return p.storage.StoreBudget(ctx, budget)
}

// GetSpentBudget implements BudgetManager.GetSpentBudget
func (p *PersistentBudgetManager) GetSpentBudget(ctx context.Context, did, context string, epoch int64) (float64, error) {
	return p.memoryManager.GetSpentBudget(ctx, did, context, epoch)
}

// ValidateBudget implements BudgetManager.ValidateBudget
func (p *PersistentBudgetManager) ValidateBudget(ctx context.Context, did, context string, epoch int64) error {
	return p.memoryManager.ValidateBudget(ctx, did, context, epoch)
}

// BudgetEnforcer provides higher-level budget enforcement logic
type BudgetEnforcer struct {
	manager    BudgetManager
	config     *ScoreConfig
	penalties  map[string]float64 // did -> reputation penalty
	mu         sync.RWMutex
}

// NewBudgetEnforcer creates a new budget enforcer
func NewBudgetEnforcer(manager BudgetManager, config *ScoreConfig) *BudgetEnforcer {
	return &BudgetEnforcer{
		manager:   manager,
		config:    config,
		penalties: make(map[string]float64),
	}
}

// EnforceVouchingLimits enforces vouching limits and applies penalties
func (e *BudgetEnforcer) EnforceVouchingLimits(ctx context.Context, did, context string, epoch int64, vouchStrength float64) error {
	// Check if DID can afford this vouch
	if err := e.manager.SpendBudget(ctx, did, context, epoch, vouchStrength); err != nil {
		// Apply penalty for attempted overspending
		e.applyOverspendingPenalty(did, vouchStrength)
		return fmt.Errorf("budget enforcement failed: %w", err)
	}
	
	return nil
}

// applyOverspendingPenalty applies a reputation penalty for overspending attempts
func (e *BudgetEnforcer) applyOverspendingPenalty(did string, attemptedAmount float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Penalty proportional to overspend amount
	penalty := attemptedAmount * 0.1 // 10% of attempted overspend
	e.penalties[did] += penalty
}

// GetReputationPenalty returns the current reputation penalty for a DID
func (e *BudgetEnforcer) GetReputationPenalty(did string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.penalties[did]
}

// ClearPenalties clears reputation penalties (called at epoch boundaries)
func (e *BudgetEnforcer) ClearPenalties() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.penalties = make(map[string]float64)
}

// GetBudgetUtilization returns budget utilization statistics
func (e *BudgetEnforcer) GetBudgetUtilization(ctx context.Context, did, context string, epoch int64) (*BudgetUtilization, error) {
	budget, err := e.manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		return nil, err
	}
	
	utilizationRate := 0.0
	if budget.TotalBudget > 0 {
		utilizationRate = budget.SpentBudget / budget.TotalBudget
	}
	
	return &BudgetUtilization{
		DID:             did,
		Context:         context,
		Epoch:           epoch,
		TotalBudget:     budget.TotalBudget,
		SpentBudget:     budget.SpentBudget,
		RemainingBudget: budget.RemainingBudget,
		UtilizationRate: utilizationRate,
		ReputationBond:  budget.ReputationBond,
		Penalty:         e.GetReputationPenalty(did),
		Timestamp:       time.Now(),
	}, nil
}

// BudgetUtilization provides budget utilization statistics
type BudgetUtilization struct {
	DID             string    `json:"did"`
	Context         string    `json:"context"`
	Epoch           int64     `json:"epoch"`
	TotalBudget     float64   `json:"total_budget"`
	SpentBudget     float64   `json:"spent_budget"`
	RemainingBudget float64   `json:"remaining_budget"`
	UtilizationRate float64   `json:"utilization_rate"` // 0.0 to 1.0
	ReputationBond  float64   `json:"reputation_bond"`
	Penalty         float64   `json:"penalty"`
	Timestamp       time.Time `json:"timestamp"`
}