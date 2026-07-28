package modifier

// Modifier stores multipliers for business revenue and speed statistics.
type Modifier struct {
	RevenueMultiplier float64 // Revenue multiplier (default 1.0)
	SpeedMultiplier   float64 // Speed multiplier (default 1.0)
}

// NewModifier creates a new Modifier instance with default multiplier values.
func NewModifier(revMult, speedMult float64) Modifier {
	if revMult <= 0 {
		revMult = 1.0
	}
	if speedMult <= 0 {
		speedMult = 1.0
	}
	return Modifier{
		RevenueMultiplier: revMult,
		SpeedMultiplier:   speedMult,
	}
}
