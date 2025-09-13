package score

import (
	"math"
)

// ExponentialDecayFunction implements DecayFunction with exponential decay
type ExponentialDecayFunction struct{}

// NewExponentialDecayFunction creates a new exponential decay function
func NewExponentialDecayFunction() DecayFunction {
	return &ExponentialDecayFunction{}
}

// ApplyDecay implements DecayFunction.ApplyDecay
// Uses exponential decay: value * (0.5)^(elapsedEpochs / halfLife)
func (d *ExponentialDecayFunction) ApplyDecay(value float64, elapsedEpochs int64, halfLife int64) float64 {
	if elapsedEpochs <= 0 || halfLife <= 0 {
		return value
	}
	
	// Calculate decay factor: (0.5)^(elapsed / halfLife)
	decayFactor := math.Pow(0.5, float64(elapsedEpochs)/float64(halfLife))
	
	return value * decayFactor
}

// ApplyInactivityDecay implements DecayFunction.ApplyInactivityDecay
// Uses exponential decay for inactivity: value * (1 - decayRate)^inactiveEpochs
func (d *ExponentialDecayFunction) ApplyInactivityDecay(value float64, inactiveEpochs int64, decayRate float64) float64 {
	if inactiveEpochs <= 0 || decayRate <= 0 {
		return value
	}
	
	// Ensure decay rate is bounded [0, 1]
	if decayRate > 1.0 {
		decayRate = 1.0
	}
	
	// Calculate decay factor: (1 - decayRate)^inactiveEpochs
	decayFactor := math.Pow(1.0-decayRate, float64(inactiveEpochs))
	
	return value * decayFactor
}

// ComputeTimeBonus implements DecayFunction.ComputeTimeBonus
// Computes bounded logarithmic growth: min(maxGrowth, log(1 + age) * activity_factor)
func (d *ExponentialDecayFunction) ComputeTimeBonus(firstActivity, lastActivity int64, currentEpoch int64, maxGrowth float64) float64 {
	if firstActivity < 0 || currentEpoch <= firstActivity {
		return 0
	}
	
	// Calculate age in epochs
	ageEpochs := float64(currentEpoch - firstActivity)
	
	// Calculate activity factor (how recently they were active)
	activityEpochs := currentEpoch - lastActivity
	activityFactor := 1.0
	if activityEpochs > 0 {
		// Reduce bonus for inactivity
		activityFactor = 1.0 / (1.0 + float64(activityEpochs)*0.1)
	}
	
	// Logarithmic growth with activity weighting
	timeBonus := math.Log(1.0+ageEpochs) * activityFactor
	
	// Apply maximum growth bound
	if timeBonus > maxGrowth {
		timeBonus = maxGrowth
	}
	
	return timeBonus
}

// LinearDecayFunction implements DecayFunction with linear decay
type LinearDecayFunction struct{}

// NewLinearDecayFunction creates a new linear decay function
func NewLinearDecayFunction() DecayFunction {
	return &LinearDecayFunction{}
}

// ApplyDecay implements DecayFunction.ApplyDecay with linear decay
// Uses linear decay: value * max(0, 1 - (elapsedEpochs / (2 * halfLife)))
func (d *LinearDecayFunction) ApplyDecay(value float64, elapsedEpochs int64, halfLife int64) float64 {
	if elapsedEpochs <= 0 || halfLife <= 0 {
		return value
	}
	
	// Linear decay factor: 1 - (elapsed / (2 * halfLife))
	// At halfLife, value becomes 0.5 * original
	// At 2 * halfLife, value becomes 0
	decayFactor := 1.0 - (float64(elapsedEpochs) / (2.0 * float64(halfLife)))
	
	// Ensure decay factor is non-negative
	if decayFactor < 0 {
		decayFactor = 0
	}
	
	return value * decayFactor
}

// ApplyInactivityDecay implements DecayFunction.ApplyInactivityDecay with linear decay
func (d *LinearDecayFunction) ApplyInactivityDecay(value float64, inactiveEpochs int64, decayRate float64) float64 {
	if inactiveEpochs <= 0 || decayRate <= 0 {
		return value
	}
	
	// Ensure decay rate is bounded
	if decayRate > 1.0 {
		decayRate = 1.0
	}
	
	// Linear decay: value * max(0, 1 - (inactiveEpochs * decayRate))
	decayFactor := 1.0 - (float64(inactiveEpochs) * decayRate)
	
	if decayFactor < 0 {
		decayFactor = 0
	}
	
	return value * decayFactor
}

// ComputeTimeBonus implements DecayFunction.ComputeTimeBonus with square root growth
func (d *LinearDecayFunction) ComputeTimeBonus(firstActivity, lastActivity int64, currentEpoch int64, maxGrowth float64) float64 {
	if firstActivity < 0 || currentEpoch <= firstActivity {
		return 0
	}
	
	// Calculate age in epochs
	ageEpochs := float64(currentEpoch - firstActivity)
	
	// Calculate activity factor
	activityEpochs := currentEpoch - lastActivity
	activityFactor := 1.0
	if activityEpochs > 0 {
		activityFactor = math.Max(0.1, 1.0-(float64(activityEpochs)*0.05))
	}
	
	// Square root growth with activity weighting
	timeBonus := math.Sqrt(ageEpochs) * activityFactor
	
	// Apply maximum growth bound
	if timeBonus > maxGrowth {
		timeBonus = maxGrowth
	}
	
	return timeBonus
}

// CustomDecayFunction allows for configurable decay parameters
type CustomDecayFunction struct {
	DecayType     string  // "exponential", "linear", "power"
	PowerExponent float64 // For power decay: value * (elapsed/halfLife)^(-exponent)
}

// NewCustomDecayFunction creates a new custom decay function
func NewCustomDecayFunction(decayType string, powerExponent float64) DecayFunction {
	return &CustomDecayFunction{
		DecayType:     decayType,
		PowerExponent: powerExponent,
	}
}

// ApplyDecay implements DecayFunction.ApplyDecay with configurable decay
func (d *CustomDecayFunction) ApplyDecay(value float64, elapsedEpochs int64, halfLife int64) float64 {
	if elapsedEpochs <= 0 || halfLife <= 0 {
		return value
	}
	
	switch d.DecayType {
	case "exponential":
		decayFactor := math.Pow(0.5, float64(elapsedEpochs)/float64(halfLife))
		return value * decayFactor
		
	case "linear":
		decayFactor := math.Max(0, 1.0-(float64(elapsedEpochs)/(2.0*float64(halfLife))))
		return value * decayFactor
		
	case "power":
		if d.PowerExponent <= 0 {
			d.PowerExponent = 1.0
		}
		ratio := float64(elapsedEpochs) / float64(halfLife)
		decayFactor := math.Pow(1.0+ratio, -d.PowerExponent)
		return value * decayFactor
		
	default:
		// Default to exponential
		decayFactor := math.Pow(0.5, float64(elapsedEpochs)/float64(halfLife))
		return value * decayFactor
	}
}

// ApplyInactivityDecay implements DecayFunction.ApplyInactivityDecay
func (d *CustomDecayFunction) ApplyInactivityDecay(value float64, inactiveEpochs int64, decayRate float64) float64 {
	if inactiveEpochs <= 0 || decayRate <= 0 {
		return value
	}
	
	if decayRate > 1.0 {
		decayRate = 1.0
	}
	
	switch d.DecayType {
	case "exponential":
		decayFactor := math.Pow(1.0-decayRate, float64(inactiveEpochs))
		return value * decayFactor
		
	case "linear":
		decayFactor := math.Max(0, 1.0-(float64(inactiveEpochs)*decayRate))
		return value * decayFactor
		
	case "power":
		if d.PowerExponent <= 0 {
			d.PowerExponent = 1.0
		}
		decayFactor := math.Pow(1.0+float64(inactiveEpochs)*decayRate, -d.PowerExponent)
		return value * decayFactor
		
	default:
		decayFactor := math.Pow(1.0-decayRate, float64(inactiveEpochs))
		return value * decayFactor
	}
}

// ComputeTimeBonus implements DecayFunction.ComputeTimeBonus
func (d *CustomDecayFunction) ComputeTimeBonus(firstActivity, lastActivity int64, currentEpoch int64, maxGrowth float64) float64 {
	if firstActivity < 0 || currentEpoch <= firstActivity {
		return 0
	}
	
	ageEpochs := float64(currentEpoch - firstActivity)
	
	// Activity factor
	activityEpochs := currentEpoch - lastActivity
	activityFactor := 1.0
	if activityEpochs > 0 {
		switch d.DecayType {
		case "exponential":
			activityFactor = math.Exp(-float64(activityEpochs) * 0.1)
		case "linear":
			activityFactor = math.Max(0.1, 1.0-(float64(activityEpochs)*0.05))
		case "power":
			if d.PowerExponent <= 0 {
				d.PowerExponent = 1.0
			}
			activityFactor = math.Pow(1.0+float64(activityEpochs)*0.1, -d.PowerExponent)
		default:
			activityFactor = 1.0 / (1.0 + float64(activityEpochs)*0.1)
		}
	}
	
	// Compute growth function
	var timeBonus float64
	switch d.DecayType {
	case "exponential":
		timeBonus = math.Log(1.0+ageEpochs) * activityFactor
	case "linear":
		timeBonus = math.Sqrt(ageEpochs) * activityFactor
	case "power":
		if d.PowerExponent <= 0 {
			d.PowerExponent = 0.5
		}
		timeBonus = math.Pow(ageEpochs, 1.0/d.PowerExponent) * activityFactor
	default:
		timeBonus = math.Log(1.0+ageEpochs) * activityFactor
	}
	
	// Apply maximum growth bound
	if timeBonus > maxGrowth {
		timeBonus = maxGrowth
	}
	
	return timeBonus
}