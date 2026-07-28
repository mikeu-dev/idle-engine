package modifier

// Modifier menyimpan pengali untuk statistik pendapatan dan kecepatan bisnis.
type Modifier struct {
	RevenueMultiplier float64 // Pengali pendapatan (default 1.0)
	SpeedMultiplier   float64 // Pengali kecepatan (default 1.0)
}

// NewModifier membuat instance Modifier baru dengan nilai pengali default.
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
