package helper

import (
	"encoding/json"
	"fmt"

	"github.com/tabed23/k8s-resource-tuner/internal/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func PrettyPrintWorkload(w models.WorkLoad) {
	out := map[string]interface{}{
		"namespace": w.Namespace,
		"name":      w.Name,
		"kind":      w.Kind,
		"labels":    w.Labels,
	}
	containers := []interface{}{}
	for _, c := range w.Containers {
		containers = append(containers, map[string]interface{}{
			"name":     c.Name,
			"requests": ResourceListToMap(c.Resources.Request),
			"limits":   ResourceListToMap(c.Resources.Limits),
		})
	}
	out["containers"] = containers
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

func ResourceListToMap(rl v1.ResourceList) map[string]string {
	m := make(map[string]string)
	for k, v := range rl {
		m[string(k)] = v.String()
	}
	return m
}

func MinMaxAvg(samples []float64) (float64, float64, float64) {
	if len(samples) == 0 {
		return 0, 0, 0
	}
	min, max, sum := samples[0], samples[0], 0.0
	for _, v := range samples {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg := sum / float64(len(samples))
	return min, max, avg
}

func BytesToMiB(b float64) float64 {
	return b / 1024 / 1024
}




func QuantityToString(q interface{}) string {
	switch val := q.(type) {
	case resource.Quantity:
		return val.String()
	case *resource.Quantity:
		if val == nil {
			return ""
		}
		return val.String()
	default:
		return fmt.Sprintf("%v", q)
	}
}

