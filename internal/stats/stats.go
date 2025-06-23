package stats

import (
	"math"
	"sort"

	"github.com/tabed23/k8s-resource-tuner/internal/models"
)

func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	s := make([]float64, len(values))
	copy(s, values)
	sort.Float64s(s)
	k := int(math.Ceil((p/100.0)*float64(len(s)))) - 1
	if k < 0 {
		k = 0
	}
	if k >= len(s) {
		k = len(s) - 1
	}
	return s[k]

}

func Avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func BuildUsageStats(containerName string, cpuSamples, memSamples []float64) models.UsageStats {
	return models.UsageStats{
		ContainerName: containerName,
		CPUSamples:    cpuSamples,
		MemSamples:    memSamples,
		CPUAvg:        Avg(cpuSamples),
		CPUP95:        Percentile(cpuSamples, 95),
		CPUP99:        Percentile(cpuSamples, 99),
		MemAvg:        Avg(memSamples),
		MemP95:        Percentile(memSamples, 95),
		MemP99:        Percentile(memSamples, 99),
	}
}