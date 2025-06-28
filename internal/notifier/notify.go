package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SlackResponse represents the response from Slack API
type SlackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	File  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"file,omitempty"`
}

// SlackAuthResponse for testing authentication
type SlackAuthResponse struct {
	OK   bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	User string `json:"user,omitempty"`
	Team string `json:"team,omitempty"`
}

// SlackUploadResponse for the new files.getUploadURLExternal API
type SlackUploadResponse struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	UploadURL string `json:"upload_url,omitempty"`
	FileID    string `json:"file_id,omitempty"`
}

// SlackCompleteResponse for completing the upload
type SlackCompleteResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	Files []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"files,omitempty"`
}

// TestSlackAuth tests if the Slack token is valid
func TestSlackAuth(slackToken string) error {
	req, err := http.NewRequest("POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return fmt.Errorf("error creating auth test request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error testing auth: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading auth response: %v", err)
	}

	var authResp SlackAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("error parsing auth response: %v", err)
	}

	if !authResp.OK {
		return fmt.Errorf("auth test failed: %s", authResp.Error)
	}

	fmt.Printf("✅ Auth test successful! User: %s, Team: %s\n", authResp.User, authResp.Team)
	return nil
}

// SendReportToSlack sends a report using the new Slack files API
func SendReportToSlack(slackToken, slackChannel, reportFilePath string) error {
	// Test authentication first
	fmt.Println("🔄 Testing Slack authentication...")
	if err := TestSlackAuth(slackToken); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Validate inputs
	if slackToken == "" {
		return fmt.Errorf("slack token cannot be empty")
	}
	if slackChannel == "" {
		return fmt.Errorf("slack channel cannot be empty")
	}
	if reportFilePath == "" {
		return fmt.Errorf("report file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(reportFilePath); os.IsNotExist(err) {
		return fmt.Errorf("report file does not exist: %s", reportFilePath)
	}

	// Get file info
	fileInfo, err := os.Stat(reportFilePath)
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}

	fileName := filepath.Base(reportFilePath)
	fileSize := fileInfo.Size()

	fmt.Printf("📤 Uploading file: %s (%.2f KB)\n", fileName, float64(fileSize)/1024)

	// Step 1: Get upload URL
	uploadURL, fileID, err := getUploadURL(slackToken, fileName, fileSize)
	if err != nil {
		return fmt.Errorf("error getting upload URL: %v", err)
	}

	// Step 2: Upload file to the URL
	if err := uploadFileToURL(uploadURL, reportFilePath); err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}

	// Step 3: Complete the upload and share to channel
	if err := completeUpload(slackToken, fileID, slackChannel, fileName); err != nil {
		return fmt.Errorf("error completing upload: %v", err)
	}

	fmt.Printf("✅ File successfully uploaded to Slack channel %s!\n", slackChannel)
	return nil
}

// getUploadURL gets the upload URL from Slack
func getUploadURL(slackToken, fileName string, fileSize int64) (string, string, error) {
	// Prepare form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("filename", fileName)
	writer.WriteField("length", fmt.Sprintf("%d", fileSize))
	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://slack.com/api/files.getUploadURLExternal", body)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response: %v", err)
	}

	var uploadResp SlackUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", "", fmt.Errorf("error parsing response: %v", err)
	}

	if !uploadResp.OK {
		return "", "", fmt.Errorf("Slack API error: %s", uploadResp.Error)
	}

	return uploadResp.UploadURL, uploadResp.FileID, nil
}

// uploadFileToURL uploads the file to the provided URL
func uploadFileToURL(uploadURL, filePath string) error {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create request
	req, err := http.NewRequest("POST", uploadURL, file)
	if err != nil {
		return fmt.Errorf("error creating upload request: %v", err)
	}

	req.Header.Set("Content-Type", "application/pdf")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	return nil
}

// completeUpload completes the upload and shares to channel
func completeUpload(slackToken, fileID, channel, fileName string) error {
	// Prepare form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)


	writer.WriteField("files", fmt.Sprintf(`[{"id":"%s","title":"%s"}]`, fileID, fileName))
	writer.WriteField("channel_id", channel)
	writer.WriteField("initial_comment", "📊 Here is the latest Kubernetes resource usage report.")
	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://slack.com/api/files.completeUploadExternal", body)
	if err != nil {
		return fmt.Errorf("error creating complete request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending complete request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading complete response: %v", err)
	}

	var completeResp SlackCompleteResponse
	if err := json.Unmarshal(respBody, &completeResp); err != nil {
		return fmt.Errorf("error parsing complete response: %v", err)
	}

	if !completeResp.OK {
		return fmt.Errorf("Slack API error: %s", completeResp.Error)
	}

	return nil
}