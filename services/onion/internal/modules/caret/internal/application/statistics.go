package application

import (
	"context"
	"errors"
	"math"
	"sort"
)

func init() {
	registerGeneralTool("utilities.statistics", (*AssistantToolExecutor).statistics)
}

const maxStatisticsValues = 5000

// statistics summarises a list of numbers exactly, so Caret never eyeballs a
// mean or median from a table.
func (e *AssistantToolExecutor) statistics(_ context.Context, input map[string]any) (map[string]any, error) {
	values := inputNumbers(input, "values", "numbers", "data")
	if len(values) == 0 {
		return nil, errors.New("give a list of numbers, e.g. {\"values\":[3,5,8]}")
	}
	if len(values) > maxStatisticsValues {
		return nil, errors.New("give at most 5,000 numbers")
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	n := float64(len(values))
	mean := sum / n
	squares := 0.0
	for _, v := range values {
		squares += (v - mean) * (v - mean)
	}
	out := map[string]any{
		"count": len(values), "sum": roundSignificant(sum, 12), "mean": roundSignificant(mean, 10),
		"median": percentile(sorted, 50), "min": sorted[0], "max": sorted[len(sorted)-1],
		"range": roundSignificant(sorted[len(sorted)-1]-sorted[0], 12),
		"p25":   percentile(sorted, 25), "p75": percentile(sorted, 75), "p90": percentile(sorted, 90),
		"population_stdev": roundSignificant(math.Sqrt(squares/n), 10),
	}
	if len(values) > 1 {
		out["sample_stdev"] = roundSignificant(math.Sqrt(squares/(n-1)), 10)
		first, last := values[0], values[len(values)-1]
		if first != 0 {
			out["change_first_to_last_percent"] = roundSignificant((last-first)/math.Abs(first)*100, 6)
		}
		if first > 0 && last > 0 {
			out["average_growth_per_step_percent"] = roundSignificant((math.Pow(last/first, 1/(n-1))-1)*100, 6)
		}
	}
	if mode, count := modeOf(sorted); count > 1 {
		out["mode"] = mode
	}
	return out, nil
}

// percentile interpolates linearly between the closest ranks.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := p / 100 * float64(len(sorted)-1)
	low := int(math.Floor(rank))
	high := int(math.Ceil(rank))
	return roundSignificant(sorted[low]+(sorted[high]-sorted[low])*(rank-float64(low)), 12)
}

func modeOf(sorted []float64) (float64, int) {
	best, bestCount, count := sorted[0], 0, 0
	for i, v := range sorted {
		if i > 0 && v == sorted[i-1] {
			count++
		} else {
			count = 1
		}
		if count > bestCount {
			best, bestCount = v, count
		}
	}
	return best, bestCount
}
