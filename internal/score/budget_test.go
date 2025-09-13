package score

import (
	"context"
	"testing"
)

func TestMemoryBudgetManager_GetBudget(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// First call should create budget
	budget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	if budget == nil {
		t.Fatal("Budget is nil")
	}
	
	if budget.DID != did {
		t.Errorf("Expected DID %s, got %s", did, budget.DID)
	}
	
	if budget.Context != context {
		t.Errorf("Expected context %s, got %s", context, budget.Context)
	}
	
	if budget.Epoch != epoch {
		t.Errorf("Expected epoch %d, got %d", epoch, budget.Epoch)
	}
	
	if budget.TotalBudget <= 0 {
		t.Errorf("Total budget should be positive, got %f", budget.TotalBudget)
	}
	
	if budget.SpentBudget != 0 {
		t.Errorf("Initial spent budget should be 0, got %f", budget.SpentBudget)
	}
	
	if budget.RemainingBudget != budget.TotalBudget {
		t.Errorf("Initial remaining budget should equal total budget")
	}
	
	// Second call should return same budget
	budget2, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("Second GetBudget failed: %v", err)
	}
	
	if budget.TotalBudget != budget2.TotalBudget {
		t.Errorf("Budget should be consistent across calls")
	}
}

func TestMemoryBudgetManager_SpendBudget(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Get initial budget
	initialBudget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	spendAmount := 5.0
	
	// Spend some budget
	err = manager.SpendBudget(ctx, did, context, epoch, spendAmount)
	if err != nil {
		t.Fatalf("SpendBudget failed: %v", err)
	}
	
	// Check updated budget
	updatedBudget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget after spend failed: %v", err)
	}
	
	if updatedBudget.SpentBudget != spendAmount {
		t.Errorf("Expected spent budget %f, got %f", spendAmount, updatedBudget.SpentBudget)
	}
	
	expectedRemaining := initialBudget.TotalBudget - spendAmount
	if updatedBudget.RemainingBudget != expectedRemaining {
		t.Errorf("Expected remaining budget %f, got %f", expectedRemaining, updatedBudget.RemainingBudget)
	}
	
	// Try to overspend
	err = manager.SpendBudget(ctx, did, context, epoch, updatedBudget.RemainingBudget+1.0)
	if err == nil {
		t.Error("SpendBudget should fail when trying to overspend")
	}
	
	// Try to spend negative amount
	err = manager.SpendBudget(ctx, did, context, epoch, -1.0)
	if err == nil {
		t.Error("SpendBudget should fail with negative amount")
	}
}

func TestMemoryBudgetManager_RefillBudget(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Get initial budget
	initialBudget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	// Spend some budget
	spendAmount := 5.0
	err = manager.SpendBudget(ctx, did, context, epoch, spendAmount)
	if err != nil {
		t.Fatalf("SpendBudget failed: %v", err)
	}
	
	// Refill with higher score
	higherScore := 50.0 // Should result in higher budget
	err = manager.RefillBudget(ctx, did, context, epoch, higherScore)
	if err != nil {
		t.Fatalf("RefillBudget failed: %v", err)
	}
	
	// Check refilled budget
	refilledBudget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget after refill failed: %v", err)
	}
	
	// Total budget should be higher now
	if refilledBudget.TotalBudget <= initialBudget.TotalBudget {
		t.Errorf("Total budget should increase after refill with higher score")
	}
	
	// Spent budget should remain the same
	if refilledBudget.SpentBudget != spendAmount {
		t.Errorf("Spent budget should remain %f after refill, got %f", spendAmount, refilledBudget.SpentBudget)
	}
	
	// Reputation bond should be updated
	expectedBond := higherScore * 0.1
	if refilledBudget.ReputationBond != expectedBond {
		t.Errorf("Expected reputation bond %f, got %f", expectedBond, refilledBudget.ReputationBond)
	}
}

func TestMemoryBudgetManager_ValidateBudget(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Create and spend some budget
	_, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	err = manager.SpendBudget(ctx, did, context, epoch, 5.0)
	if err != nil {
		t.Fatalf("SpendBudget failed: %v", err)
	}
	
	// Validate should pass
	err = manager.ValidateBudget(ctx, did, context, epoch)
	if err != nil {
		t.Errorf("ValidateBudget failed: %v", err)
	}
}

func TestBudgetEnforcer_EnforceVouchingLimits(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	enforcer := NewBudgetEnforcer(manager, config)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Get initial budget to know the limit
	budget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	// Should be able to vouch within budget
	validAmount := budget.RemainingBudget / 2
	err = enforcer.EnforceVouchingLimits(ctx, did, context, epoch, validAmount)
	if err != nil {
		t.Errorf("EnforceVouchingLimits should succeed within budget: %v", err)
	}
	
	// Should fail to vouch beyond remaining budget
	excessAmount := budget.RemainingBudget + 10.0
	err = enforcer.EnforceVouchingLimits(ctx, did, context, epoch, excessAmount)
	if err == nil {
		t.Error("EnforceVouchingLimits should fail when exceeding budget")
	}
	
	// Check that penalty was applied
	penalty := enforcer.GetReputationPenalty(did)
	if penalty <= 0 {
		t.Error("Reputation penalty should be applied for overspending attempt")
	}
}

func TestBudgetEnforcer_GetBudgetUtilization(t *testing.T) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	enforcer := NewBudgetEnforcer(manager, config)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Get initial budget
	budget, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	
	// Spend half the budget
	spendAmount := budget.TotalBudget / 2
	err = manager.SpendBudget(ctx, did, context, epoch, spendAmount)
	if err != nil {
		t.Fatalf("SpendBudget failed: %v", err)
	}
	
	// Get utilization
	utilization, err := enforcer.GetBudgetUtilization(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetBudgetUtilization failed: %v", err)
	}
	
	if utilization == nil {
		t.Fatal("Utilization is nil")
	}
	
	expectedRate := 0.5 // 50% utilization
	if utilization.UtilizationRate < expectedRate-0.01 || utilization.UtilizationRate > expectedRate+0.01 {
		t.Errorf("Expected utilization rate %f, got %f", expectedRate, utilization.UtilizationRate)
	}
	
	if utilization.SpentBudget != spendAmount {
		t.Errorf("Expected spent budget %f, got %f", spendAmount, utilization.SpentBudget)
	}
}

func BenchmarkMemoryBudgetManager_GetBudget(b *testing.B) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:benchmark"
	context := "benchmark_context"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		epoch := int64(i)
		_, err := manager.GetBudget(ctx, did, context, epoch)
		if err != nil {
			b.Fatalf("GetBudget failed: %v", err)
		}
	}
}

func BenchmarkMemoryBudgetManager_SpendBudget(b *testing.B) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	
	ctx := context.Background()
	did := "did:key:benchmark"
	context := "benchmark_context"
	epoch := int64(1000)
	
	// Pre-create budget
	_, err := manager.GetBudget(ctx, did, context, epoch)
	if err != nil {
		b.Fatalf("GetBudget failed: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create new budget for each iteration to avoid depletion
		testEpoch := epoch + int64(i)
		_, err := manager.GetBudget(ctx, did, context, testEpoch)
		if err != nil {
			b.Fatalf("GetBudget failed: %v", err)
		}
		
		err = manager.SpendBudget(ctx, did, context, testEpoch, 1.0)
		if err != nil {
			b.Fatalf("SpendBudget failed: %v", err)
		}
	}
}

func BenchmarkBudgetEnforcer_EnforceVouchingLimits(b *testing.B) {
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	manager := NewMemoryBudgetManager(config, dataProvider)
	enforcer := NewBudgetEnforcer(manager, config)
	
	ctx := context.Background()
	did := "did:key:benchmark"
	context := "benchmark_context"
	epoch := int64(1000)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create new budget for each iteration
		testEpoch := epoch + int64(i)
		err := enforcer.EnforceVouchingLimits(ctx, did, context, testEpoch, 1.0)
		if err != nil {
			b.Fatalf("EnforceVouchingLimits failed: %v", err)
		}
	}
}