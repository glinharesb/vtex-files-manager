package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/glinharesb/vtex-files-manager/pkg/auth"
	"github.com/glinharesb/vtex-files-manager/pkg/client"
	"github.com/glinharesb/vtex-files-manager/pkg/vtexcli"
	"github.com/spf13/cobra"
)

var (
	uploadMethod string
	skipConfirm  bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "Upload a single file to VTEX",
	Long: `Upload a single image or file to your VTEX account.

Authentication:
  Uses VTEX CLI session. Run 'vtex login' first if not logged in.

Supported file types:
  - Universal (both methods): jpg, jpeg, png, gif, svg, webp
  - CMS only: bmp, pdf, txt, json, xml, css, js
Maximum file size: 5MB

Upload Methods:
  graphql: Official GraphQL API - URLs: account.vtexassets.com/assets/.../uuid___hash.ext
  cms:     Legacy CMS FilePicker - URLs: account.vtexassets.com/arquivos/filename.ext

Note: You must specify the --method flag. There is no default value.

Examples:
  vtex-files-manager upload image.jpg -m cms
  vtex-files-manager upload logo.png -m graphql -y
  vtex-files-manager upload banner.jpg -m cms -v`,
	Args: cobra.ExactArgs(1),
	RunE: runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&uploadMethod, "method", "m", "", "upload method: graphql or cms (required)")
	uploadCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "skip confirmation prompt")
}

func runUpload(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Validate method is specified
	if uploadMethod == "" {
		return fmt.Errorf("--method flag is required (must be 'graphql' or 'cms')")
	}

	// Validate method value
	if uploadMethod != "graphql" && uploadMethod != "cms" {
		return fmt.Errorf("invalid method: %s (must be 'graphql' or 'cms')", uploadMethod)
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

	// Create authenticator
	authenticator := auth.NewAuthenticator(session.Token)

	// Get file info for display
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}

	// Build destination URL
	fileName := filepath.Base(filePath)
	var destURL string
	if uploadMethod == "cms" {
		// Use URL encoding for filenames with spaces or special characters
		encodedFileName := url.PathEscape(fileName)
		destURL = fmt.Sprintf("https://%s.vtexassets.com/arquivos/%s", session.Account, encodedFileName)
	} else {
		destURL = fmt.Sprintf("https://%s.vtexassets.com/assets/.../[generated]", session.Account)
	}

	// Check if file exists (only for CMS method)
	fileExists := false
	if uploadMethod == "cms" {
		cmsClient := client.NewCMSFilePickerClient(session.Account, session.Workspace, authenticator, verbose)
		exists, err := cmsClient.CheckFileExists(fileName)
		if err != nil && verbose {
			fmt.Printf("Warning: Could not check if file exists: %v\n", err)
		}
		fileExists = exists
	}

	// Display upload info
	infoColor := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	infoColor.Println("=== VTEX File Upload ===")
	fmt.Printf("Account:       %s\n", session.Account)
	fmt.Printf("Workspace:     %s\n", session.Workspace)
	fmt.Printf("User:          %s\n", session.Login)
	fmt.Printf("Method:        %s\n", uploadMethod)
	fmt.Printf("File:          %s (%.2f KB)\n", fileName, float64(fileInfo.Size())/1024)
	fmt.Printf("Destination:   %s\n", destURL)

	// Show warning if file exists
	if fileExists {
		color.Yellow("\n⚠️  WARNING: File already exists and will be OVERWRITTEN!")
	}

	fmt.Println()

	// Ask for confirmation unless --yes flag is set
	if !skipConfirm {
		promptMsg := "Proceed with upload?"
		if fileExists {
			promptMsg = "File exists. Overwrite?"
		}
		if !askConfirmation(promptMsg) {
			color.Yellow("Upload cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Upload file based on method
	var result *client.UploadResult
	if uploadMethod == "cms" {
		// Use CMS FilePicker client
		cmsClient := client.NewCMSFilePickerClient(session.Account, session.Workspace, authenticator, verbose)
		result, err = cmsClient.UploadFile(filePath, true)
	} else {
		// Use GraphQL client (default)
		graphqlClient := client.NewGraphQLClient(session.Account, session.Workspace, authenticator, verbose)
		result, err = graphqlClient.UploadFile(filePath, true)
	}

	if err != nil {
		errorColor := color.New(color.FgRed, color.Bold)
		errorColor.Printf("\n✗ Upload failed: %v\n", err)
		return err
	}

	// Print success message
	successColor := color.New(color.FgGreen, color.Bold)
	fmt.Println()
	successColor.Println("✓ Upload successful!")
	fmt.Printf("File URL: %s\n", result.FileURL)
	fmt.Println()

	return nil
}
