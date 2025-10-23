package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/glinharesb/vtex-files-manager/pkg/auth"
	"github.com/glinharesb/vtex-files-manager/pkg/logger"
	"github.com/schollz/progressbar/v3"
)

// FilePickerUploadResponse represents the response from FilePicker upload
type FilePickerUploadResponse struct {
	FileNameInserted string `json:"fileNameInserted"`
	Mensagem         string `json:"mensagem"`
}

// CMSFilePickerClient represents a VTEX CMS FilePicker client
type CMSFilePickerClient struct {
	account       string
	workspace     string
	authenticator *auth.Authenticator
	httpClient    *http.Client
	verbose       bool
	requestToken  string
}

// NewCMSFilePickerClient creates a new VTEX CMS FilePicker client
func NewCMSFilePickerClient(account, workspace string, authenticator *auth.Authenticator, verbose bool) *CMSFilePickerClient {
	return &CMSFilePickerClient{
		account:       account,
		workspace:     workspace,
		authenticator: authenticator,
		httpClient: &http.Client{
			Timeout: 5 * 60 * 1000000000, // 5 minutes
		},
		verbose: verbose,
	}
}

// getRequestToken fetches the requestToken from the CMS admin page
func (c *CMSFilePickerClient) getRequestToken() error {
	// URL to get the upload page that contains the requestToken
	url := fmt.Sprintf("https://%s.vtexcommercestable.com.br/admin/a/PortalManagement/AddFile?fileType=images", c.account)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	c.authenticator.AddAuthHeaders(req)

	if c.verbose {
		fmt.Printf("Fetching requestToken from: %s\n", url)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Check for authentication errors
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return fmt.Errorf("authentication failed (HTTP %d): your VTEX session has expired. Please run 'vtex login' and try again", resp.StatusCode)
		}
		if resp.StatusCode == 302 {
			return fmt.Errorf("authentication failed (redirect): your VTEX session has expired. Please run 'vtex login' and try again")
		}
		return fmt.Errorf("failed to fetch upload page with status %d: %s", resp.StatusCode, string(body))
	}

	if c.verbose {
		maxLen := 2000
		if len(body) < maxLen {
			maxLen = len(body)
		}
		fmt.Printf("HTML Response (first %d chars):\n%s\n", maxLen, string(body[:maxLen]))
	}

	// Extract requestToken from HTML
	// Looking for: <input type="hidden" id="fileUploadRequestToken" value="TOKEN_HERE" />
	re := regexp.MustCompile(`id="fileUploadRequestToken"\s+value="([^"]+)"`)
	matches := re.FindSubmatch(body)

	if len(matches) < 2 {
		// Try alternative pattern
		re = regexp.MustCompile(`value="([^"]+)"\s+id="fileUploadRequestToken"`)
		matches = re.FindSubmatch(body)
	}

	if len(matches) < 2 {
		// Try searching for any input with fileUploadRequestToken
		re = regexp.MustCompile(`fileUploadRequestToken[^>]*value="([^"]+)"`)
		matches = re.FindSubmatch(body)
	}

	if len(matches) < 2 {
		if c.verbose {
			fmt.Printf("Full HTML Response:\n%s\n", string(body))
		}
		return fmt.Errorf("authentication failed: could not obtain upload token. Your VTEX session may have expired. Please run 'vtex login' and try again")
	}

	c.requestToken = string(matches[1])

	if c.verbose {
		fmt.Printf("RequestToken obtained: %s\n", c.requestToken)
	}

	return nil
}

// UploadFile uploads a single file using CMS FilePicker
func (c *CMSFilePickerClient) UploadFile(filePath string, showProgress bool) (*UploadResult, error) {
	result := &UploadResult{
		FileName: filepath.Base(filePath),
	}

	// Validate file
	if err := ValidateFile(filePath); err != nil {
		result.Error = err
		return result, err
	}

	// ALWAYS get a fresh requestToken before each upload
	// The token has a very short lifespan (seconds) and must be obtained immediately before upload
	if err := c.getRequestToken(); err != nil {
		result.Error = fmt.Errorf("failed to get requestToken: %w", err)
		return result, result.Error
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file: %w", err)
		return result, result.Error
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		result.Error = fmt.Errorf("failed to get file info: %w", err)
		return result, result.Error
	}

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileName := filepath.Base(filePath)

	// Add requestToken field
	if err := writer.WriteField("requestToken", c.requestToken); err != nil {
		result.Error = fmt.Errorf("failed to write requestToken field: %w", err)
		return result, result.Error
	}

	// Add the file itself (field name must be "FileData" with capital D)
	// Set Content-Type based on file extension
	ext := filepath.Ext(filePath)
	mimeType := GetMIMEType(ext)

	// Create part with explicit Content-Type
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="FileData"; filename="%s"`, fileName)}
	h["Content-Type"] = []string{mimeType}

	part, err := writer.CreatePart(h)
	if err != nil {
		result.Error = fmt.Errorf("failed to create form file part: %w", err)
		return result, result.Error
	}

	// Copy file content with optional progress bar
	var fileReader io.Reader = file
	if showProgress {
		bar := progressbar.DefaultBytes(
			fileInfo.Size(),
			fmt.Sprintf("Uploading %s", fileName),
		)
		fileReader = io.TeeReader(file, bar)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		result.Error = fmt.Errorf("failed to copy file content: %w", err)
		return result, result.Error
	}

	if err := writer.Close(); err != nil {
		result.Error = fmt.Errorf("failed to close multipart writer: %w", err)
		return result, result.Error
	}

	// Upload via FilePicker
	fileURL, err := c.uploadFilePicker(body, writer.FormDataContentType(), fileName)
	if err != nil {
		result.Error = err

		// Log failed upload
		logger.LogUpload(logger.UploadLogEntry{
			Timestamp: time.Now(),
			File:      fileName,
			Path:      filePath,
			Size:      fileInfo.Size(),
			Method:    "cms",
			Account:   c.account,
			Workspace: c.workspace,
			Status:    "failed",
			Error:     err.Error(),
		})

		return result, result.Error
	}

	result.FileURL = fileURL
	result.Success = true

	// Log successful upload
	logger.LogUpload(logger.UploadLogEntry{
		Timestamp: time.Now(),
		File:      fileName,
		Path:      filePath,
		Size:      fileInfo.Size(),
		Method:    "cms",
		Account:   c.account,
		Workspace: c.workspace,
		Status:    "success",
		URL:       fileURL,
	})

	return result, nil
}

// uploadFilePicker performs the FilePicker upload request
func (c *CMSFilePickerClient) uploadFilePicker(body *bytes.Buffer, contentType, fileName string) (string, error) {
	// Build FilePicker endpoint URL
	url := fmt.Sprintf("https://%s.vtexcommercestable.com.br/admin/a/FilePicker/UploadFile", c.account)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// Add authentication headers
	c.authenticator.AddAuthHeaders(req)

	if c.verbose {
		fmt.Printf("FilePicker Endpoint: %s\n", url)
		fmt.Printf("Auth method: %s\n", c.authenticator.GetMethodName())
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("Response Status: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(respBody))
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Check for authentication errors
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return "", fmt.Errorf("authentication failed (HTTP %d): your VTEX session has expired. Please run 'vtex login' and try again", resp.StatusCode)
		}
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse JSON response
	var uploadResp FilePickerUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	// Check if upload was successful
	if uploadResp.FileNameInserted == "" {
		return "", fmt.Errorf("upload failed: %s", uploadResp.Mensagem)
	}

	// Build the file URL for /arquivos path
	// FilePicker uploads go to: https://{account}.vtexassets.com/arquivos/{filename}
	// Use URL encoding for filenames with spaces or special characters
	encodedFileName := neturl.PathEscape(uploadResp.FileNameInserted)
	fileURL := fmt.Sprintf("https://%s.vtexassets.com/arquivos/%s", c.account, encodedFileName)

	if c.verbose {
		fmt.Printf("Upload successful! Message: %s\n", uploadResp.Mensagem)
		fmt.Printf("File URL: %s\n", fileURL)
	}

	return fileURL, nil
}

// CheckFileExists verifies if a file already exists in VTEX FilePicker
func (c *CMSFilePickerClient) CheckFileExists(fileName string) (bool, error) {
	url := fmt.Sprintf("https://%s.vtexcommercestable.com.br/admin/a/FilePicker/FileExists?changedFileName=", c.account)

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add filename field
	if err := writer.WriteField(fileName, fileName); err != nil {
		return false, fmt.Errorf("failed to write field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return false, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "*/*")

	// Add authentication headers
	c.authenticator.AddAuthHeaders(req)

	if c.verbose {
		fmt.Printf("Checking if file exists: %s\n", fileName)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("FileExists response: %s\n", string(respBody))
	}

	// Parse JSON response
	var result map[string]string
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	// File exists if the map contains the filename as key
	_, exists := result[fileName]
	return exists, nil
}
