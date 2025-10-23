package vtexcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SessionData represents VTEX CLI session data from session.json
type SessionData struct {
	Account     string `json:"account"`
	Login       string `json:"login"`
	Token       string `json:"token"`
	LastAccount string `json:"lastAccount"`
}

// WorkspaceData represents VTEX CLI workspace data from workspace.json
type WorkspaceData struct {
	CurrentWorkspace string `json:"currentWorkspace"`
	LastWorkspace    string `json:"lastWorkspace"`
}

// VTEXSession represents a complete VTEX CLI session
type VTEXSession struct {
	Account   string
	Login     string
	Token     string
	Workspace string
}

// getVTEXSessionPath returns the path to VTEX CLI session directory
func getVTEXSessionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".vtex", "session"), nil
}

// LoadSession loads the current VTEX CLI session
func LoadSession() (*VTEXSession, error) {
	sessionPath, err := getVTEXSessionPath()
	if err != nil {
		return nil, err
	}

	// Check if session directory exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no VTEX session found. Please run 'vtex login' first")
	}

	// Read session.json
	sessionFile := filepath.Join(sessionPath, "session.json")
	sessionBytes, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w. Please run 'vtex login' first", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	// Read workspace.json
	workspaceFile := filepath.Join(sessionPath, "workspace.json")
	workspaceBytes, err := os.ReadFile(workspaceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace file: %w. Please run 'vtex login' first", err)
	}

	var workspaceData WorkspaceData
	if err := json.Unmarshal(workspaceBytes, &workspaceData); err != nil {
		return nil, fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Validate required fields
	if sessionData.Account == "" {
		return nil, fmt.Errorf("no account found in session. Please run 'vtex login' first")
	}
	if sessionData.Token == "" {
		return nil, fmt.Errorf("no token found in session. Please run 'vtex login' first")
	}
	if workspaceData.CurrentWorkspace == "" {
		return nil, fmt.Errorf("no workspace found in session. Please run 'vtex use <workspace>' first")
	}

	return &VTEXSession{
		Account:   sessionData.Account,
		Login:     sessionData.Login,
		Token:     sessionData.Token,
		Workspace: workspaceData.CurrentWorkspace,
	}, nil
}

// ValidateToken performs basic validation on the authentication token
// Returns an error if the token appears to be invalid
func (s *VTEXSession) ValidateToken() error {
	if s.Token == "" {
		return fmt.Errorf("no authentication token found")
	}

	// VTEX tokens are typically long strings (base64/JWT format)
	// A token shorter than 10 characters is likely invalid or a placeholder
	if len(s.Token) < 10 {
		return fmt.Errorf("authentication token appears to be invalid (too short)")
	}

	return nil
}
