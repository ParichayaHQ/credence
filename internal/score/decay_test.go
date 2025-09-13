package score

import (
	"math"
	"testing"
)

func TestExponentialDecayFunction_ApplyDecay(t *testing.T) {
	decay := NewExponentialDecayFunction()
	
	tests := []struct {
		name           string
		value          float64
		elapsedEpochs  int64
		halfLife       int64
		expectedResult float64
		tolerance      float64
	}{
		{
			name:           "no decay at zero elapsed",
			value:          100.0,
			elapsedEpochs:  0,
			halfLife:       10,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
		{
			name:           "half value at half life",
			value:          100.0,
			elapsedEpochs:  10,
			halfLife:       10,
			expectedResult: 50.0,
			tolerance:      0.001,
		},
		{
			name:           "quarter value at double half life",
			value:          100.0,
			elapsedEpochs:  20,
			halfLife:       10,
			expectedResult: 25.0,
			tolerance:      0.001,
		},
		{
			name:           "negative elapsed returns original",
			value:          100.0,
			elapsedEpochs:  -5,
			halfLife:       10,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
		{
			name:           "zero half life returns original",
			value:          100.0,
			elapsedEpochs:  10,
			halfLife:       0,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decay.ApplyDecay(tt.value, tt.elapsedEpochs, tt.halfLife)
			if math.Abs(result-tt.expectedResult) > tt.tolerance {
				t.Errorf("Expected %f, got %f", tt.expectedResult, result)
			}
		})
	}
}

func TestExponentialDecayFunction_ApplyInactivityDecay(t *testing.T) {
	decay := NewExponentialDecayFunction()
	
	tests := []struct {
		name           string
		value          float64
		inactiveEpochs int64
		decayRate      float64
		expectedResult float64
		tolerance      float64
	}{
		{
			name:           "no inactivity no decay",
			value:          100.0,
			inactiveEpochs: 0,
			decayRate:      0.1,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
		{
			name:           "10% decay rate for 1 epoch",
			value:          100.0,
			inactiveEpochs: 1,
			decayRate:      0.1,
			expectedResult: 90.0,
			tolerance:      0.001,
		},
		{
			name:           "compound decay over multiple epochs",
			value:          100.0,
			inactiveEpochs: 2,
			decayRate:      0.1,
			expectedResult: 81.0,
			tolerance:      0.001,
		},
		{
			name:           "zero decay rate returns original",
			value:          100.0,
			inactiveEpochs: 10,
			decayRate:      0.0,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
		{
			name:           "decay rate > 1 is capped at 1",
			value:          100.0,
			inactiveEpochs: 1,
			decayRate:      1.5,
			expectedResult: 0.0,
			tolerance:      0.001,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decay.ApplyInactivityDecay(tt.value, tt.inactiveEpochs, tt.decayRate)
			if math.Abs(result-tt.expectedResult) > tt.tolerance {
				t.Errorf("Expected %f, got %f", tt.expectedResult, result)
			}
		})
	}
}

func TestExponentialDecayFunction_ComputeTimeBonus(t *testing.T) {
	decay := NewExponentialDecayFunction()
	
	tests := []struct {
		name          string
		firstActivity int64
		lastActivity  int64
		currentEpoch  int64
		maxGrowth     float64
		expectPositive bool
		expectCapped   bool
	}{
		{
			name:          "basic time bonus",
			firstActivity: 0,
			lastActivity:  90,
			currentEpoch:  100,
			maxGrowth:     50.0,
			expectPositive: true,
			expectCapped:   false,
		},
		{
			name:          "capped at max growth",
			firstActivity: 0,
			lastActivity:  95,
			currentEpoch:  10000,
			maxGrowth:     10.0,
			expectPositive: true,
			expectCapped:   true,
		},
		{
			name:          "no bonus for future first activity",
			firstActivity: 200,
			lastActivity:  250,
			currentEpoch:  100,
			maxGrowth:     50.0,
			expectPositive: false,
			expectCapped:   false,
		},
		{
			name:          "inactivity reduces bonus",
			firstActivity: 0,
			lastActivity:  50,
			currentEpoch:  100,
			maxGrowth:     50.0,
			expectPositive: true,
			expectCapped:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decay.ComputeTimeBonus(tt.firstActivity, tt.lastActivity, tt.currentEpoch, tt.maxGrowth)
			
			if tt.expectPositive && result <= 0 {
				t.Errorf("Expected positive result, got %f", result)
			}
			
			if !tt.expectPositive && result > 0 {
				t.Errorf("Expected non-positive result, got %f", result)
			}
			
			if tt.expectCapped && result > tt.maxGrowth+0.001 {
				t.Errorf("Expected result to be capped at %f, got %f", tt.maxGrowth, result)
			}
		})
	}
}

func TestLinearDecayFunction_ApplyDecay(t *testing.T) {
	decay := NewLinearDecayFunction()
	
	tests := []struct {
		name           string
		value          float64
		elapsedEpochs  int64
		halfLife       int64
		expectedResult float64
		tolerance      float64
	}{
		{
			name:           "no decay at zero elapsed",
			value:          100.0,
			elapsedEpochs:  0,
			halfLife:       10,
			expectedResult: 100.0,
			tolerance:      0.001,
		},
		{
			name:           "half value at half life",
			value:          100.0,
			elapsedEpochs:  10,
			halfLife:       10,
			expectedResult: 50.0,
			tolerance:      0.001,
		},
		{
			name:           "zero at double half life",
			value:          100.0,
			elapsedEpochs:  20,
			halfLife:       10,
			expectedResult: 0.0,
			tolerance:      0.001,
		},
		{
			name:           "clamped to zero beyond double half life",
			value:          100.0,
			elapsedEpochs:  30,
			halfLife:       10,
			expectedResult: 0.0,
			tolerance:      0.001,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decay.ApplyDecay(tt.value, tt.elapsedEpochs, tt.halfLife)
			if math.Abs(result-tt.expectedResult) > tt.tolerance {
				t.Errorf("Expected %f, got %f", tt.expectedResult, result)
			}
		})
	}
}

func TestCustomDecayFunction(t *testing.T) {
	// Test exponential decay type
	expDecay := NewCustomDecayFunction("exponential", 1.0)
	result1 := expDecay.ApplyDecay(100.0, 10, 10)
	expected1 := 50.0
	if math.Abs(result1-expected1) > 0.001 {
		t.Errorf("Exponential decay: expected %f, got %f", expected1, result1)
	}
	
	// Test linear decay type
	linDecay := NewCustomDecayFunction("linear", 1.0)
	result2 := linDecay.ApplyDecay(100.0, 10, 10)
	expected2 := 50.0
	if math.Abs(result2-expected2) > 0.001 {
		t.Errorf("Linear decay: expected %f, got %f", expected2, result2)
	}
	
	// Test power decay type
	powDecay := NewCustomDecayFunction("power", 2.0)
	result3 := powDecay.ApplyDecay(100.0, 10, 10)
	expected3 := 100.0 / math.Pow(2.0, 2.0) // Should be 25.0
	if math.Abs(result3-expected3) > 0.001 {
		t.Errorf("Power decay: expected %f, got %f", expected3, result3)
	}
	
	// Test unknown type defaults to exponential
	unknownDecay := NewCustomDecayFunction("unknown", 1.0)
	result4 := unknownDecay.ApplyDecay(100.0, 10, 10)
	expected4 := 50.0
	if math.Abs(result4-expected4) > 0.001 {
		t.Errorf("Unknown decay type: expected %f, got %f", expected4, result4)
	}
}

func BenchmarkExponentialDecay_ApplyDecay(b *testing.B) {
	decay := NewExponentialDecayFunction()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decay.ApplyDecay(100.0, 10, 20)
	}
}

func BenchmarkExponentialDecay_ApplyInactivityDecay(b *testing.B) {
	decay := NewExponentialDecayFunction()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decay.ApplyInactivityDecay(100.0, 5, 0.1)
	}
}

func BenchmarkExponentialDecay_ComputeTimeBonus(b *testing.B) {
	decay := NewExponentialDecayFunction()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decay.ComputeTimeBonus(0, 90, 100, 50.0)
	}
}

func BenchmarkLinearDecay_ApplyDecay(b *testing.B) {
	decay := NewLinearDecayFunction()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decay.ApplyDecay(100.0, 10, 20)
	}
}

func BenchmarkCustomDecay_PowerDecay(b *testing.B) {
	decay := NewCustomDecayFunction("power", 2.0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decay.ApplyDecay(100.0, 10, 20)
	}
}