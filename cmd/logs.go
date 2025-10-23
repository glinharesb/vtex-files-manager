package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/glinharesb/vtex-files-manager/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	logsLimit  int
	logsStatus string
	logsMethod string
	logsClear  bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View upload history logs",
	Long: `View the history of all file uploads made with vtex-files-manager.

Logs are stored in JSONL format at:
  - Linux:   ~/.local/state/vtex-files-manager/uploads.jsonl
  - macOS:   ~/Library/Application Support/vtex-files-manager/uploads.jsonl
  - Windows: %LOCALAPPDATA%\vtex-files-manager\uploads.jsonl

Each log entry includes: timestamp, file name, size, upload method,
account, workspace, status (success/failed), resulting URL, and error message if failed.

Examples:
  vtex-files-manager logs
  vtex-files-manager logs --limit 10
  vtex-files-manager logs --status failed
  vtex-files-manager logs --method cms
  vtex-files-manager logs --clear`,
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().IntVarP(&logsLimit, "limit", "l", 50, "maximum number of entries to display")
	logsCmd.Flags().StringVarP(&logsStatus, "status", "s", "", "filter by status: success or failed")
	logsCmd.Flags().StringVarP(&logsMethod, "method", "m", "", "filter by upload method: graphql or cms")
	logsCmd.Flags().BoolVarP(&logsClear, "clear", "c", false, "clear all logs (requires confirmation)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Handle --clear flag
	if logsClear {
		return clearLogsWithConfirmation()
	}

	// Get log file path
	logPath, err := logger.GetLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Read all logs
	entries, err := logger.ReadLogs()
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	if len(entries) == 0 {
		color.Yellow("No upload logs found.")
		fmt.Printf("\nLog file location: %s\n", logPath)
		return nil
	}

	// Apply filters
	filteredEntries := filterEntries(entries, logsStatus, logsMethod)

	if len(filteredEntries) == 0 {
		color.Yellow("No entries match the specified filters.")
		fmt.Printf("\nTotal entries in log: %d\n", len(entries))
		fmt.Printf("Log file location: %s\n", logPath)
		return nil
	}

	// Apply limit (show most recent entries)
	displayEntries := filteredEntries
	if logsLimit > 0 && len(filteredEntries) > logsLimit {
		displayEntries = filteredEntries[len(filteredEntries)-logsLimit:]
	}

	// Print header
	headerColor := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	headerColor.Println("=== VTEX Upload Logs ===")
	fmt.Printf("Showing %d of %d entries", len(displayEntries), len(entries))
	if logsStatus != "" || logsMethod != "" {
		fmt.Printf(" (filtered)")
	}
	fmt.Println()
	fmt.Printf("Log file: %s\n", logPath)
	fmt.Println()

	// Print entries
	for i, entry := range displayEntries {
		printLogEntry(i+1, entry)
	}

	// Print summary
	printLogsSummary(filteredEntries)

	return nil
}

func filterEntries(entries []logger.UploadLogEntry, status, method string) []logger.UploadLogEntry {
	filtered := []logger.UploadLogEntry{}

	for _, entry := range entries {
		// Filter by status
		if status != "" && entry.Status != status {
			continue
		}

		// Filter by method
		if method != "" && entry.Method != method {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func printLogEntry(index int, entry logger.UploadLogEntry) {
	// Format timestamp
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Status with color
	var statusStr string
	if entry.Status == "success" {
		statusStr = color.GreenString("✓ SUCCESS")
	} else {
		statusStr = color.RedString("✗ FAILED")
	}

	// Entry header
	fmt.Printf("[%d] %s | %s\n", index, timestamp, statusStr)

	// File info
	sizeKB := float64(entry.Size) / 1024
	fmt.Printf("    File:      %s (%.2f KB)\n", entry.File, sizeKB)

	// Path if available
	if entry.Path != "" {
		fmt.Printf("    Path:      %s\n", entry.Path)
	}

	// Upload details
	fmt.Printf("    Method:    %s\n", entry.Method)
	fmt.Printf("    Account:   %s\n", entry.Account)
	fmt.Printf("    Workspace: %s\n", entry.Workspace)

	// URL or Error
	if entry.Status == "success" && entry.URL != "" {
		fmt.Printf("    URL:       %s\n", entry.URL)
	} else if entry.Status == "failed" && entry.Error != "" {
		fmt.Printf("    Error:     %s\n", color.RedString(entry.Error))
	}

	fmt.Println()
}

func printLogsSummary(entries []logger.UploadLogEntry) {
	successCount := 0
	failedCount := 0
	cmsCount := 0
	graphqlCount := 0

	for _, entry := range entries {
		if entry.Status == "success" {
			successCount++
		} else {
			failedCount++
		}

		if entry.Method == "cms" {
			cmsCount++
		} else if entry.Method == "graphql" {
			graphqlCount++
		}
	}

	summaryColor := color.New(color.FgCyan, color.Bold)
	summaryColor.Println("=== Summary ===")
	fmt.Printf("Total:         %d uploads\n", len(entries))
	color.Green("Successful:    %d", successCount)
	if failedCount > 0 {
		color.Red("Failed:        %d", failedCount)
	} else {
		fmt.Printf("Failed:        %d\n", failedCount)
	}
	fmt.Printf("CMS uploads:   %d\n", cmsCount)
	fmt.Printf("GraphQL:       %d\n", graphqlCount)
	fmt.Println()
}

func clearLogsWithConfirmation() error {
	// Read current logs to show count
	entries, err := logger.ReadLogs()
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	if len(entries) == 0 {
		color.Yellow("No logs to clear.")
		return nil
	}

	// Get log file path for display
	logPath, _ := logger.GetLogPath()

	// Show warning
	color.Yellow("\n⚠️  WARNING: This will permanently delete all upload logs!")
	fmt.Printf("Log file: %s\n", logPath)
	fmt.Printf("Total entries: %d\n\n", len(entries))

	// Ask for confirmation
	if !askConfirmation("Are you sure you want to clear all logs?") {
		color.Yellow("Operation cancelled.")
		return nil
	}

	// Clear logs
	if err := logger.ClearLogs(); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	color.Green("\n✓ Logs cleared successfully!")
	return nil
}
