package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/glinharesb/vtex-files-manager/pkg/auth"
	"github.com/glinharesb/vtex-files-manager/pkg/logger"
	"github.com/schollz/progressbar/v3"
)

// GraphQLClient represents a VTEX GraphQL API client
type GraphQLClient struct {
	account       string
	workspace     string
	authenticator *auth.Authenticator
	httpClient    *http.Client
	verbose       bool
}

// GraphQLUploadResult represents the result of a GraphQL file upload
type GraphQLUploadResult struct {
	Data struct {
		UploadFile struct {
			FileURL  string `json:"fileUrl"`
			MimeType string `json:"mimetype"`
			Encoding string `json:"encoding"`
		} `json:"uploadFile"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
}

// NewGraphQLClient creates a new VTEX GraphQL API client
func NewGraphQLClient(account, workspace string, authenticator *auth.Authenticator, verbose bool) *GraphQLClient {
	return &GraphQLClient{
		account:       account,
		workspace:     workspace,
		authenticator: authenticator,
		httpClient: &http.Client{
			Timeout: 5 * 60 * 1000000000, // 5 minutes
		},
		verbose: verbose,
	}
}

// UploadFile uploads a single file using GraphQL mutation
func (c *GraphQLClient) UploadFile(filePath string, showProgress bool) (*UploadResult, error) {
	result := &UploadResult{
		FileName: filepath.Base(filePath),
	}

	// Validate file
	if err := ValidateFile(filePath); err != nil {
		result.Error = err
		return result, err
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

	// Prepare GraphQL multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 1. Add operations (GraphQL query)
	operations := map[string]interface{}{
		"query": `mutation uploadFile($file: Upload!, $bucket: String) {
			uploadFile(file: $file, bucket: $bucket) {
				fileUrl
				mimetype
				encoding
			}
		}`,
		"variables": map[string]interface{}{
			"file":   nil, // Will be mapped from the file part
			"bucket": "images",
		},
	}

	operationsJSON, err := json.Marshal(operations)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal operations: %w", err)
		return result, result.Error
	}

	if err := writer.WriteField("operations", string(operationsJSON)); err != nil {
		result.Error = fmt.Errorf("failed to write operations field: %w", err)
		return result, result.Error
	}

	// 2. Add map (file mapping)
	fileMap := map[string][]string{
		"0": {"variables.file"},
	}

	mapJSON, err := json.Marshal(fileMap)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal map: %w", err)
		return result, result.Error
	}

	if err := writer.WriteField("map", string(mapJSON)); err != nil {
		result.Error = fmt.Errorf("failed to write map field: %w", err)
		return result, result.Error
	}

	// 3. Add the file itself with proper Content-Type
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="0"; filename="%s"`, filepath.Base(filePath))}

	// Set Content-Type based on file extension
	ext := filepath.Ext(filePath)
	mimeType := GetMIMEType(ext)
	h["Content-Type"] = []string{mimeType}

	part, err := writer.CreatePart(h)
	if err != nil {
		result.Error = fmt.Errorf("failed to create form file: %w", err)
		return result, result.Error
	}

	// Copy file content with optional progress bar
	var fileReader io.Reader = file
	if showProgress {
		bar := progressbar.DefaultBytes(
			fileInfo.Size(),
			fmt.Sprintf("Uploading %s", filepath.Base(filePath)),
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

	// Upload with GraphQL
	fileURL, err := c.uploadGraphQL(body, writer.FormDataContentType())
	if err != nil {
		result.Error = err

		// Log failed upload
		logger.LogUpload(logger.UploadLogEntry{
			Timestamp: time.Now(),
			File:      filepath.Base(filePath),
			Path:      filePath,
			Size:      fileInfo.Size(),
			Method:    "graphql",
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
		File:      filepath.Base(filePath),
		Path:      filePath,
		Size:      fileInfo.Size(),
		Method:    "graphql",
		Account:   c.account,
		Workspace: c.workspace,
		Status:    "success",
		URL:       fileURL,
	})

	return result, nil
}

// uploadGraphQL performs the GraphQL upload request
func (c *GraphQLClient) uploadGraphQL(body *bytes.Buffer, contentType string) (string, error) {
	// Build GraphQL endpoint URL
	// Use the account-specific endpoint
	url := fmt.Sprintf("https://%s.myvtex.com/_v/private/graphql/v1", c.account)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	// Add authentication headers
	c.authenticator.AddAuthHeaders(req)

	if c.verbose {
		fmt.Printf("GraphQL Endpoint: %s\n", url)
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

	// Parse GraphQL response
	var gqlResult GraphQLUploadResult
	if err := json.Unmarshal(respBody, &gqlResult); err != nil {
		return "", fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResult.Errors) > 0 {
		errMsg := gqlResult.Errors[0].Message
		return "", fmt.Errorf("GraphQL error: %s", errMsg)
	}

	// Get file URL from response
	fileURL := gqlResult.Data.UploadFile.FileURL
	if fileURL == "" {
		return "", fmt.Errorf("no fileUrl in response")
	}

	if c.verbose {
		fmt.Printf("Upload successful! File URL: %s\n", fileURL)
	}

	return fileURL, nil
}
