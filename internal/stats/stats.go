package stats

import (
	"math"
	"sort"
)

type Stats struct {
	Count  int     `json:"count"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"std_dev"`
	CV     float64 `json:"cv"`
	P10    float64 `json:"p10"`
	P90    float64 `json:"p90"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
}

type Comparison struct {
	GainPercent float64 `json:"gain_percent"`
	GainP10     float64 `json:"gain_p10"`
	GainP90     float64 `json:"gain_p90"`
	Conclusive  bool    `json:"conclusive"`
	Overlap     float64 `json:"overlap"`
}

func Calculate(values []float64) Stats {
	if len(values) == 0 {
		return Stats{}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	stats := Stats{
		Count:  n,
		Min:    sorted[0],
		Max:    sorted[n-1],
		Median: percentile(sorted, 50),
		P10:    percentile(sorted, 10),
		P90:    percentile(sorted, 90),
		P95:    percentile(sorted, 95),
		P99:    percentile(sorted, 99),
	}

	// Mean
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	stats.Mean = sum / float64(n)

	// Standard deviation
	sumSq := 0.0
	for _, v := range sorted {
		diff := v - stats.Mean
		sumSq += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSq / float64(n))

	// Coefficient of variation (%)
	if stats.Mean != 0 {
		stats.CV = (stats.StdDev / stats.Mean) * 100
	}

	return stats
}

func Compare(baseline, optimized []float64, alternate bool) Comparison {
	if len(baseline) == 0 || len(optimized) == 0 {
		return Comparison{}
	}

	baselineStats := Calculate(baseline)
	optimizedStats := Calculate(optimized)

	// Main gain calculation using medians (robust)
	gainPercent := 0.0
	if baselineStats.Median > 0 {
		gainPercent = ((baselineStats.Median - optimizedStats.Median) / baselineStats.Median) * 100
	}

	comp := Comparison{
		GainPercent: gainPercent,
	}

	// Calculate pairwise gains if alternating (more accurate P10/P90)
	if alternate && len(baseline) == len(optimized) {
		pairwiseGains := make([]float64, len(baseline))
		for i := range baseline {
			if baseline[i] > 0 {
				pairwiseGains[i] = ((baseline[i] - optimized[i]) / baseline[i]) * 100
			}
		}
		pairStats := Calculate(pairwiseGains)
		comp.GainP10 = pairStats.P10
		comp.GainP90 = pairStats.P90
	} else {
		// Estimate from distribution overlap
		comp.GainP10 = gainPercent - (baselineStats.CV + optimizedStats.CV) / 2
		comp.GainP90 = gainPercent + (baselineStats.CV + optimizedStats.CV) / 2
	}

	// Calculate overlap between distributions
	comp.Overlap = calculateOverlap(baseline, optimized)

	// Determine if result is conclusive
	// Conclusive if: low CV, low overlap, consistent direction
	avgCV := (baselineStats.CV + optimizedStats.CV) / 2
	comp.Conclusive = avgCV < 15 && comp.Overlap < 0.3 && 
		((comp.GainP10 > 0 && comp.GainP90 > 0) || (comp.GainP10 < 0 && comp.GainP90 < 0))

	return comp
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	idx := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))

	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}

	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

func calculateOverlap(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	statsA := Calculate(a)
	statsB := Calculate(b)

	// Simple overlap estimation based on range intersection
	minA, maxA := statsA.P10, statsA.P90
	minB, maxB := statsB.P10, statsB.P90

	overlapStart := math.Max(minA, minB)
	overlapEnd := math.Min(maxA, maxB)

	if overlapStart >= overlapEnd {
		return 0 // No overlap
	}

	overlapRange := overlapEnd - overlapStart
	totalRange := math.Max(maxA, maxB) - math.Min(minA, minB)

	if totalRange == 0 {
		return 1
	}

	return overlapRange / totalRange
}

// TrimmedMean calculates mean after removing top/bottom percentage
func TrimmedMean(values []float64, trimPercent float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	trimCount := int(float64(len(sorted)) * trimPercent / 100.0)
	if trimCount*2 >= len(sorted) {
		return Calculate(values).Median
	}

	trimmed := sorted[trimCount : len(sorted)-trimCount]
	sum := 0.0
	for _, v := range trimmed {
		sum += v
	}
	return sum / float64(len(trimmed))
}
