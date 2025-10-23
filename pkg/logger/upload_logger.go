package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/adrg/xdg"
)

const logFileName = "vtex-files-manager/uploads.jsonl"

// UploadLogEntry represents a single upload operation in the log
type UploadLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	File      string    `json:"file"`
	Path      string    `json:"path,omitempty"`
	Size      int64     `json:"size"`
	Method    string    `json:"method"` // "cms" or "graphql"
	Account   string    `json:"account"`
	Workspace string    `json:"workspace"`
	Status    string    `json:"status"` // "success" or "failed"
	URL       string    `json:"url,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// LogUpload appends an upload entry to the log file
func LogUpload(entry UploadLogEntry) error {
	// Get log file path (creates parent directories if needed)
	logPath, err := xdg.StateFile(logFileName)
	if err != nil {
		return err
	}

	// Open file in append mode
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add timestamp if not present
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Serialize to JSON and write line
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Append newline to create JSONL format
	_, err = file.Write(append(data, '\n'))
	return err
}

// ReadLogs reads all upload log entries from the log file
func ReadLogs() ([]UploadLogEntry, error) {
	// Search for log file
	logPath, err := xdg.SearchStateFile(logFileName)
	if err != nil {
		// No logs file exists yet
		return []UploadLogEntry{}, nil
	}

	// Read file line by line
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []UploadLogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry UploadLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			// Skip invalid lines
			continue
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

// GetLogPath returns the path to the log file
func GetLogPath() (string, error) {
	return xdg.StateFile(logFileName)
}

// ClearLogs removes the log file
func ClearLogs() error {
	logPath, err := xdg.SearchStateFile(logFileName)
	if err != nil {
		// File doesn't exist, nothing to clear
		return nil
	}

	return os.Remove(logPath)
}
