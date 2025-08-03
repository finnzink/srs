package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Version information - injected at build time
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

const (
	GitHubRepo        = "finnzink/srs"
	UpdateCheckFile   = ".srs_last_check"
	CheckInterval     = 24 * time.Hour // Check once per day
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func printVersion() {
	fmt.Printf("srs version %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Built: %s\n", BuildTime)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func shouldCheckForUpdates() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	checkFile := filepath.Join(homeDir, UpdateCheckFile)
	
	// Check if file exists and when it was last modified
	info, err := os.Stat(checkFile)
	if err != nil {
		// File doesn't exist, we should check
		return true
	}

	// Check if enough time has passed
	return time.Since(info.ModTime()) > CheckInterval
}

func updateLastCheckTime() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	checkFile := filepath.Join(homeDir, UpdateCheckFile)
	
	// Create or update the file with current timestamp
	file, err := os.Create(checkFile)
	if err != nil {
		return
	}
	defer file.Close()
	
	file.WriteString(time.Now().Format(time.RFC3339))
}

func normalizeVersion(v string) string {
	// Remove 'v' prefix if present
	return strings.TrimPrefix(v, "v")
}

func checkForUpdates() {
	if !shouldCheckForUpdates() {
		return
	}

	// Update the check time first to avoid multiple concurrent checks
	updateLastCheckTime()

	latestVersion, err := getLatestVersion()
	if err != nil {
		// Silently fail - don't bother users with network issues
		return
	}

	currentVersion := normalizeVersion(Version)
	latest := normalizeVersion(latestVersion)

	// Simple version comparison - for now just compare strings
	// (This works for semver but could be improved with proper semver parsing)
	if latest != currentVersion && latest > currentVersion {
		fmt.Printf("📦 Update available: v%s → v%s\n", currentVersion, latest)
		fmt.Printf("Run 'srs update' to upgrade\n\n")
	}
}