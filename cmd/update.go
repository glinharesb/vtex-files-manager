package cmd

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	forceUpdate bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update vfm to the latest version",
	Long: `Update vfm to the latest version available on GitHub Releases.

This command checks for new versions and automatically downloads and installs
the latest binary for your platform.

Examples:
  vfm update                 # Update to latest version
  vfm update --check         # Only check for updates, don't install
  vfm update --force         # Force update even if same version`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "only check for updates, don't install")
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "force update even if same version")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Get current version from build-time variable
	currentVersion := version
	if currentVersion == "" {
		currentVersion = "dev"
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Checking for updates...\n", cyan("ℹ"))
	fmt.Printf("Current version: %s\n", currentVersion)

	// Create self-updater
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{}, // No filters, use all assets
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	// Check for latest release
	latest, found, err := updater.DetectLatest("glinharesb/vtex-files-manager")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !found {
		fmt.Printf("%s No releases found\n", yellow("⚠"))
		return nil
	}

	latestVersion := latest.Version.String()
	fmt.Printf("Latest version:  %s\n", latestVersion)

	// Check if update is needed
	if currentVersion == latestVersion && !forceUpdate {
		fmt.Printf("\n%s You're already on the latest version!\n", green("✓"))
		return nil
	}

	// Compare versions using semver
	if currentVersion != "dev" && !forceUpdate {
		currentSemVer, err := semver.Parse(currentVersion)
		if err == nil {
			// If current version >= latest version, no update needed
			if currentSemVer.GTE(latest.Version) {
				fmt.Printf("\n%s You're already on the latest version (or newer)!\n", green("✓"))
				return nil
			}
		}
	}

	// If check-only mode, stop here
	if checkOnly {
		fmt.Printf("\n%s New version available: %s → %s\n", cyan("ℹ"), currentVersion, latestVersion)
		fmt.Printf("Run 'vfm update' to install the latest version\n")
		return nil
	}

	// Confirm update
	if !forceUpdate {
		fmt.Printf("\n%s Update available: %s → %s\n", yellow("⚠"), currentVersion, latestVersion)
		fmt.Printf("Do you want to update? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Update cancelled")
			return nil
		}
	}

	// Perform update
	fmt.Printf("\n%s Downloading update...\n", cyan("⬇"))

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := updater.UpdateTo(latest, exe); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	fmt.Printf("%s Successfully updated to version %s!\n", green("✓"), latestVersion)
	fmt.Printf("\nRelease notes: %s\n", latest.ReleaseNotes)

	return nil
}
