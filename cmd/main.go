package main

import (
	"fmt"

	"github.com/tabed23/k8s-resource-tuner/internal/k8s"
	"github.com/tabed23/k8s-resource-tuner/internal/prometheus"
	"github.com/tabed23/k8s-resource-tuner/internal/report"
)


func main() {
	clientSet, err := k8s.InitKubeClient()
	if err != nil {
		panic(err)
	}
	prom := prometheus.NewPromClient("http://localhost:9090")
	reportData, err := report.GenrateReport(clientSet, prom, "")
	if err != nil {
		panic(err)
	}
	reportPDF, err := report.PDFReport(reportData,"" )
	if err != nil {
		panic(err)
	}
	fmt.Printf("Report generated successfully: %s\n", reportPDF)

}
