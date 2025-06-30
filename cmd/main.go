package main

import (
	"fmt"
	"time"

	"github.com/tabed23/k8s-resource-tuner/internal/k8s"
	"github.com/tabed23/k8s-resource-tuner/internal/models"
	"github.com/tabed23/k8s-resource-tuner/internal/notifier"
	"github.com/tabed23/k8s-resource-tuner/internal/prometheus"
	"github.com/tabed23/k8s-resource-tuner/internal/report"
)

const (
	slackToken   = ""
	slackChannel = ""
)

var (
	namespaces = []string{"test"}
)

func main() {
	clientSet, err := k8s.InitKubeClient()
	if err != nil {
		panic(err)
	}
	prom := prometheus.NewPromClient("http://localhost:9090")
	var allEntries []models.ReportEntry
	var allSummaries string

	for _, ns := range namespaces {
		reportData, err := report.GenrateReport(clientSet, prom, ns)
		if err != nil {
			fmt.Printf("Error generating report for %s: %v\n", ns, err)
			continue
		}
		allEntries = append(allEntries, reportData.Entries...)
		allSummaries += fmt.Sprintf("Namespace: %s\n%s\n", ns, reportData.Summary)
	}

	combinedReport := models.Report{
		Timestamp: time.Now(),
		Entries:   allEntries,
		Summary:   allSummaries,
	}

	reportPDF, err := report.PDFReport(combinedReport, "ALL_NAMESPACES")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Combined report generated successfully: %s\n", reportPDF)
	if err := notifier.SendReportToSlack(slackToken, slackChannel, reportPDF); err != nil {
		panic(err)
	}

}
