package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// MaxFileSize is the maximum file size allowed (5MB)
	MaxFileSize = 5 * 1024 * 1024
)

// UploadResult represents the result of a file upload operation
type UploadResult struct {
	FileName string
	FileURL  string
	Success  bool
	Error    error
}

// ValidExtensions contains file extensions validated by testing
// against VTEX's actual APIs (tested on 2025-10-23)
//
// All formats below work with both CMS FilePicker and GraphQL methods
// unless otherwise noted.
var ValidExtensions = map[string]bool{
	// Universal image formats (work with both CMS and GraphQL)
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".svg":  true,
	".webp": true,

	// Additional formats supported only by CMS FilePicker
	// (GraphQL returns "Invalid file format" for these)
	".bmp":  true, // CMS only
	".pdf":  true, // CMS only
	".txt":  true, // CMS only
	".json": true, // CMS only
	".css":  true, // CMS only
	".js":   true, // CMS only
	".xml":  true, // CMS only
}

// GetMIMEType returns the MIME type for a given file extension
func GetMIMEType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}

// ValidateFile validates that a file exists and meets requirements for upload
func ValidateFile(filePath string) error {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	// Check if it's a file (not a directory)
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Check file size
	if fileInfo.Size() > MaxFileSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes / 5MB)",
			fileInfo.Size(), MaxFileSize)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty: %s", filePath)
	}

	// Check file extension (case-insensitive)
	ext := strings.ToLower(filepath.Ext(filePath))
	if !ValidExtensions[ext] {
		return fmt.Errorf("unsupported file type: %s (images: jpg, jpeg, png, gif, svg, webp, bmp; docs: pdf, txt, json, xml; web: css, js)", ext)
	}

	return nil
}
