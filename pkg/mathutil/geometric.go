package mathutil

import "math"

// CalculateMaxLevelsAffordable calculates the maximum upgrade levels that can be purchased with the given balance.
// Returns the number of purchasable levels and the total cost required.
func CalculateMaxLevelsAffordable(baseCost, multiplier float64, currentLevel int, balance float64) (int, float64) {
	costNext := baseCost * math.Pow(multiplier, float64(currentLevel))
	if balance < costNext {
		return 0, 0
	}

	var n float64
	if math.Abs(multiplier-1.0) < 1e-9 {
		n = math.Floor(balance / costNext)
	} else {
		// Use logarithmic closed-form formula for geometric series
		n = math.Floor(math.Log(1.0+(balance*(multiplier-1.0))/costNext) / math.Log(multiplier))
	}

	levels := int(n)
	if levels < 0 {
		levels = 0
	}

	// Calculate the exact total cost
	var totalCost float64
	if math.Abs(multiplier-1.0) < 1e-9 {
		totalCost = costNext * float64(levels)
	} else {
		totalCost = costNext * (math.Pow(multiplier, float64(levels)) - 1.0) / (multiplier - 1.0)
	}

	return levels, totalCost
}
