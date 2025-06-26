package report

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/tabed23/k8s-resource-tuner/internal/helper"
	"github.com/tabed23/k8s-resource-tuner/internal/k8s"
	"github.com/tabed23/k8s-resource-tuner/internal/models"
	"github.com/tabed23/k8s-resource-tuner/internal/prometheus"
	"github.com/tabed23/k8s-resource-tuner/internal/recommendation"
	"github.com/tabed23/k8s-resource-tuner/internal/stats"
	"k8s.io/client-go/kubernetes"
)

func GenrateReport(clientset *kubernetes.Clientset, prom *prometheus.PromClient, namespace string)(models.Report, error) {

	worloads, err := k8s.ListDeployments(clientset, namespace)
	if err != nil {
		panic(err)
	}
	for _, w := range worloads {
		helper.PrettyPrintWorkload(w)
	}
	if len(worloads) == 0 {
		fmt.Println("No deployments found in namespace anubis")
	}

	var reportEntries []models.ReportEntry
	var summary string

	end := time.Now()
	start := end.Add(-1 * time.Hour) // Last hour
	step := "60"

	for _, w := range worloads {
		var statsList []models.UsageStats
		var recommendations []models.Recommendation

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

			rec := recommendation.RecommendFromStats(usageStats)
			statsList = append(statsList, usageStats)
			recommendations = append(recommendations, rec)

		}
		reportEntries = append(reportEntries, models.ReportEntry{
			Workload:       w,
			Recommendation: recommendations,
		})

		summary += fmt.Sprintf("Workload: %s\n", w.Name)

	}

	return models.Report{
		Timestamp: time.Now(),
		Entries:   reportEntries,
		Summary:   summary,
	}, nil

}


func PDFReport(reportData models.Report, namespace string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Resource Usage and Recommendations", true)
	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)

	pdf.Cell(200, 10, fmt.Sprintf("Report generated on: %s", reportData.Timestamp.Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(200, 10, "Summary:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, reportData.Summary, "", "", false)

	pdf.Ln(10)

	for _, entry := range reportData.Entries {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(200, 10, fmt.Sprintf("Workload: %s (%s)", entry.Workload.Name, entry.Workload.Namespace))
		pdf.Ln(6)

		for _, rec := range entry.Recommendation {
			pdf.SetFont("Arial", "", 12)
			pdf.Cell(200, 10, fmt.Sprintf("Container: %s", rec.ContainerName))
			pdf.Ln(6)

			cpuRequest := rec.RecommendedRequest.Request["cpu"]
			cpuLimit := rec.RecommendedLimit.Limits["cpu"]
			memRequest := rec.RecommendedRequest.Request["memory"]
			memLimit := rec.RecommendedLimit.Limits["memory"]

			pdf.Cell(200, 10, fmt.Sprintf("Recommended CPU Request: %s, Recommended CPU Limit: %s", cpuRequest.String(), cpuLimit.String()))
			pdf.Ln(6)
			pdf.Cell(200, 10, fmt.Sprintf("Recommended Memory Request: %s, Recommended Memory Limit: %s", memRequest.String(), memLimit.String()))
			pdf.Ln(6)
		}
		pdf.Ln(8)
	}

	reportFilename := fmt.Sprintf("resource_report_%s_%s.pdf", namespace,reportData.Timestamp.Format("20060102_150405"))
	err := pdf.OutputFileAndClose(reportFilename)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF report: %v", err)
	}

	return reportFilename, nil
}