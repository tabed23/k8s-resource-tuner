package main

import (
	"fmt"
	"time"

	"github.com/tabed23/k8s-resource-tuner/internal/helper"
	"github.com/tabed23/k8s-resource-tuner/internal/k8s"
	"github.com/tabed23/k8s-resource-tuner/internal/models"
	"github.com/tabed23/k8s-resource-tuner/internal/prometheus"
	"github.com/tabed23/k8s-resource-tuner/internal/recommendation"
	"github.com/tabed23/k8s-resource-tuner/internal/stats"
)

func main() {
	clientSet, err := k8s.InitKubeClient()
	if err != nil {
		panic(err)
	}
	worloads, err := k8s.ListDeployments(clientSet, "vyro-env")
	if err != nil {
		panic(err)
	}
	for _, w := range worloads {
		helper.PrettyPrintWorkload(w)
	}
	if len(worloads) == 0 {
		fmt.Println("No deployments found in namespace anubis")
		return
	}
	prom := prometheus.NewPromClient("http://localhost:9090")
	end := time.Now()
	start := end.Add(-1 * time.Hour) // Last hour
	step := "60"

	for _, w := range worloads {
		fmt.Printf("\n----\nQuerying Prometheus for Deployment: %s\n", w.Name)
		for _, container := range w.Containers {
			cpuVals, err := prom.QueryCpu(w.Namespace, w.Name, start, end, step)
			if err != nil {
				fmt.Printf("Error querying CPU for container %s: %v\n", container.Name, err)
				continue
			}
			memVals, err := prom.QueryMemory(w.Namespace, w.Name, start, end, step)
			if err != nil {
				fmt.Printf("Error querying Memory for container %s: %v\n", container.Name, err)
				continue
			}

			usageStats := models.UsageStats{
				ContainerName: container.Name,
				CPUSamples:    cpuVals,
				MemSamples:    memVals,
				CPUAvg:        stats.Avg(cpuVals),
				CPUP95:        stats.Percentile(cpuVals, 95),
				CPUP99:        stats.Percentile(cpuVals, 99),
				MemAvg:        stats.Avg(memVals),
				MemP95:        stats.Percentile(memVals, 95),
				MemP99:        stats.Percentile(memVals, 99),
			}

			fmt.Printf("\nContainer: %s\n", usageStats.ContainerName)
			fmt.Printf("  CPU Avg:    %.4f cores\n", usageStats.CPUAvg)
			fmt.Printf("  CPU p95:    %.4f cores\n", usageStats.CPUP95)
			fmt.Printf("  CPU p99:    %.4f cores\n", usageStats.CPUP99)
			fmt.Printf("  Mem Avg:    %.2f MiB\n", helper.BytesToMiB(usageStats.MemAvg))
			fmt.Printf("  Mem p95:    %.2f MiB\n", helper.BytesToMiB(usageStats.MemP95))
			fmt.Printf("  Mem p99:    %.2f MiB\n", helper.BytesToMiB(usageStats.MemP99))

			rec := recommendation.RecommendFromStats(usageStats)

			fmt.Printf("\n>>> Recommendation for container '%s':\n", rec.ContainerName)
			fmt.Printf("  CPU Request: %s\n", helper.QuantityToString(rec.RecommendedRequest.Request["cpu"]))
			fmt.Printf("  CPU Limit:   %s\n", helper.QuantityToString(rec.RecommendedLimit.Limits["cpu"]))
			fmt.Printf("  Mem Request: %s\n", helper.QuantityToString(rec.RecommendedRequest.Request["memory"]))
			fmt.Printf("  Mem Limit:   %s\n", helper.QuantityToString(rec.RecommendedLimit.Limits["memory"]))
			fmt.Printf("  Reason:      %s\n", rec.Reason)


		}
	}

}
