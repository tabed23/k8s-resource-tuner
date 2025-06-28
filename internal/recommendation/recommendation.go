package recommendation

import (
	"fmt"
	"math"

	"github.com/tabed23/k8s-resource-tuner/internal/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func roundMillicores(cpu float64) string {
	mc := math.Ceil(cpu * 1000)
	return fmt.Sprintf("%.0fm", mc)
}

func roundMiB(memBytes float64) string {
	mb := math.Ceil(memBytes / 1024 / 1024)
	return fmt.Sprintf("%.0fMi", mb)
}

func resourceMustParse(val string) resource.Quantity {
	q, err := resource.ParseQuantity(val)
	if err != nil {
		return resource.MustParse("1")
	}
	return q
}

func RecommendFromStats(stats models.UsageStats) models.Recommendation {
	cpuRequest := stats.CPUP95
	cpuLimit := stats.CPUP99

	memRequest := stats.MemP95
	memLimit := stats.MemP99

	req := v1.ResourceList{
		v1.ResourceCPU:    resourceMustParse(roundMillicores(cpuRequest)),
		v1.ResourceMemory: resourceMustParse(roundMiB(memRequest)),
	}
	lim := v1.ResourceList{
		v1.ResourceCPU:    resourceMustParse(roundMillicores(cpuLimit)),
		v1.ResourceMemory: resourceMustParse(roundMiB(memLimit)),
	}
	return models.Recommendation{
		ContainerName:      stats.ContainerName,
		RecommendedRequest: models.ResourceConfig{Request: req},
		RecommendedLimit:   models.ResourceConfig{Limits: lim},
		Reason:             "Based on observed p95 (requests) and p99 (limits) from recent metrics",
	}
}
