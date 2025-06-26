package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tabed23/k8s-resource-tuner/internal/models"
)

func SendReportToSlack(slackWebhookURL string, reportData models.Report, reportFilePath string) error {
	message := fmt.Sprintf("Resource Usage Report for %s\nTimestamp: %s\n\n", reportData.Summary, reportData.Timestamp.Format("2006-01-02 15:04:05"))
	message += fmt.Sprintf("Download the full report here: %s", reportFilePath)

	payload := map[string]interface{}{
		"text": message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	req, err := http.NewRequest("POST", slackWebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to Slack: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack API returned non-200 status: %d, %s", resp.StatusCode, string(body))
	}

	return nil
}
