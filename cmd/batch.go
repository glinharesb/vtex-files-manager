package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/glinharesb/vtex-files-manager/pkg/auth"
	"github.com/glinharesb/vtex-files-manager/pkg/client"
	"github.com/glinharesb/vtex-files-manager/pkg/vtexcli"
	"github.com/spf13/cobra"
)

var (
	concurrency      int
	recursive        bool
	batchMethod      string
	batchSkipConfirm bool
)

var batchCmd = &cobra.Command{
	Use:   "batch [directory]",
	Short: "Upload multiple files from a directory",
	Long: `Upload all image files from a directory to your VTEX account.

Authentication:
  Uses VTEX CLI session. Run 'vtex login' first if not logged in.

Supported file types:
  - Universal (both methods): jpg, jpeg, png, gif, svg, webp
  - CMS only: bmp, pdf, txt, json, xml, css, js
Maximum file size: 5MB per file

Upload Methods:
  graphql: Official GraphQL API - URLs: account.vtexassets.com/assets/.../uuid___hash.ext
  cms:     Legacy CMS FilePicker - URLs: account.vtexassets.com/arquivos/filename.ext

Note: You must specify the --method flag. There is no default value.

Examples:
  vtex-files-manager batch ./images -m cms
  vtex-files-manager batch ./assets -m graphql -c 5 -y
  vtex-files-manager batch ./photos -m cms -r
  vtex-files-manager batch ./images -m graphql -c 3 -v`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().StringVarP(&batchMethod, "method", "m", "", "upload method: graphql or cms (required)")
	batchCmd.Flags().IntVarP(&concurrency, "concurrent", "c", 3, "number of concurrent uploads")
	batchCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recursively search subdirectories")
	batchCmd.Flags().BoolVarP(&batchSkipConfirm, "yes", "y", false, "skip confirmation prompt")
}

func runBatch(cmd *cobra.Command, args []string) error {
	directory := args[0]

	// Validate method is specified
	if batchMethod == "" {
		return fmt.Errorf("--method flag is required (must be 'graphql' or 'cms')")
	}

	// Validate method value
	if batchMethod != "graphql" && batchMethod != "cms" {
		return fmt.Errorf("invalid method: %s (must be 'graphql' or 'cms')", batchMethod)
	}

	// Load VTEX CLI session
	session, err := vtexcli.LoadSession()
	if err != nil {
		return err
	}

	// Validate token before proceeding
	if err := session.ValidateToken(); err != nil {
		return fmt.Errorf("authentication failed: %w. Please run 'vtex login' and try again", err)
	}

	// Find all image files
	files, err := findImageFiles(directory, recursive)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		color.Yellow("No image files found in %s", directory)
		return nil
	}

	// Calculate total size
	var totalSize int64
	for _, f := range files {
		info, err := os.Stat(f)
		if err == nil {
			totalSize += info.Size()
		}
	}

	// Create authenticator (needed for both checking and uploading)
	authenticator := auth.NewAuthenticator(session.Token)

	// Check which files already exist (only for CMS method)
	existingFiles := []string{}
	if batchMethod == "cms" {
		cmsClient := client.NewCMSFilePickerClient(session.Account, session.Workspace, authenticator, verbose)

		for _, f := range files {
			fileName := filepath.Base(f)
			exists, err := cmsClient.CheckFileExists(fileName)
			if err != nil && verbose {
				fmt.Printf("Warning: Could not check if %s exists: %v\n", fileName, err)
			}
			if exists {
				existingFiles = append(existingFiles, fileName)
			}
		}
	}

	// Print upload info
	infoColor := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	infoColor.Println("=== VTEX Batch Upload ===")
	fmt.Printf("Account:       %s\n", session.Account)
	fmt.Printf("Workspace:     %s\n", session.Workspace)
	fmt.Printf("User:          %s\n", session.Login)
	fmt.Printf("Method:        %s\n", batchMethod)
	fmt.Printf("Directory:     %s\n", directory)
	fmt.Printf("Files found:   %d (%.2f MB total)\n", len(files), float64(totalSize)/(1024*1024))
	fmt.Printf("Concurrency:   %d workers\n", concurrency)
	fmt.Println()

	// Show file list (max 10 files)
	fmt.Println("Files to upload:")
	displayLimit := 10
	for i, f := range files {
		if i >= displayLimit {
			fmt.Printf("  ... (%d more)\n", len(files)-displayLimit)
			break
		}
		info, _ := os.Stat(f)
		fmt.Printf("  %d. %s (%.2f KB)\n", i+1, filepath.Base(f), float64(info.Size())/1024)
	}
	fmt.Println()

	// Show warning if files already exist
	if len(existingFiles) > 0 {
		color.Yellow("⚠️  WARNING: %d file(s) already exist and will be OVERWRITTEN:", len(existingFiles))
		displayLimit := 5
		for i, f := range existingFiles {
			if i >= displayLimit {
				fmt.Printf("  ... and %d more\n", len(existingFiles)-displayLimit)
				break
			}
			fmt.Printf("  • %s\n", f)
		}
		fmt.Println()
	}

	// Ask for confirmation unless --yes flag is set
	if !batchSkipConfirm {
		promptMsg := "Proceed with upload?"
		if len(existingFiles) > 0 {
			promptMsg = fmt.Sprintf("%d file(s) will be overwritten. Continue?", len(existingFiles))
		}
		if !askConfirmation(promptMsg) {
			color.Yellow("Upload cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Upload files concurrently
	results := uploadFilesWithConcurrency(session.Account, session.Workspace, authenticator, files, concurrency, batchMethod)

	// Print summary
	printBatchSummary(results)

	return nil
}

func findImageFiles(directory string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := filepath.Ext(path)
				if client.ValidExtensions[ext] {
					files = append(files, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				if client.ValidExtensions[ext] {
					files = append(files, filepath.Join(directory, entry.Name()))
				}
			}
		}
	}

	return files, nil
}

func uploadFilesWithConcurrency(account, workspace string, authenticator *auth.Authenticator, files []string, concurrency int, method string) []*client.UploadResult {
	results := make([]*client.UploadResult, 0, len(files))
	var resultsMutex sync.Mutex

	// Create channels
	fileChan := make(chan string, len(files))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Create client for this worker based on method
			var uploadFunc func(string, bool) (*client.UploadResult, error)

			if method == "cms" {
				cmsClient := client.NewCMSFilePickerClient(account, workspace, authenticator, verbose)
				uploadFunc = cmsClient.UploadFile
			} else {
				graphqlClient := client.NewGraphQLClient(account, workspace, authenticator, verbose)
				uploadFunc = graphqlClient.UploadFile
			}

			for filePath := range fileChan {
				fmt.Printf("[Worker %d] Uploading: %s\n", workerID+1, filepath.Base(filePath))

				result, err := uploadFunc(filePath, false)
				if err != nil {
					color.Red("  ✗ Failed: %v", err)
				} else {
					color.Green("  ✓ Success: %s", result.FileURL)
				}

				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()

				// Small delay to avoid rate limiting
				time.Sleep(500 * time.Millisecond)
			}
		}(i)
	}

	// Send files to workers
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Wait for all workers to finish
	wg.Wait()

	return results
}

func printBatchSummary(results []*client.UploadResult) {
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("=== Upload Summary ===")
	fmt.Printf("Total files:     %d\n", len(results))
	color.Green("Successful:      %d", successCount)
	if failureCount > 0 {
		color.Red("Failed:          %d", failureCount)
	} else {
		fmt.Printf("Failed:          %d\n", failureCount)
	}
	fmt.Println()

	if failureCount > 0 {
		color.Yellow("Failed uploads:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  • %s: %v\n", result.FileName, result.Error)
			}
		}
		fmt.Println()
	}
}
