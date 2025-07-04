package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// BuildInfo contains version information set at build time
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available      bool   `json:"available"`
	LatestVersion  string `json:"latestVersion"`
	CurrentVersion string `json:"currentVersion"`
	ReleaseURL     string `json:"releaseURL"`
	ReleaseNotes   string `json:"releaseNotes"`
	Error          string `json:"error,omitempty"`
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	PreRelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

// VersionFile represents the VERSION.json file structure
type VersionFile struct {
	Version string `json:"version"`
}

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	Date      = "unknown"
	GoVersion = "unknown"
)

// GetBuildInfo returns the current build information
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version:   getVersion(),
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}

	// Try to get additional info from debug.ReadBuildInfo
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if info.GoVersion == "unknown" {
			info.GoVersion = buildInfo.GoVersion
		}

		// Extract VCS info if available
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if info.Commit == "unknown" {
					info.Commit = setting.Value
				}
			case "vcs.time":
				if info.Date == "unknown" {
					info.Date = setting.Value
				}
			}
		}
	}

	return info
}

// getVersion returns the version, prioritizing build-time version over file
func getVersion() string {
	// If we have a build-time version (not dev), use it
	if Version != "dev" && Version != "unknown" && Version != "" {
		return Version
	}

	// For development, try to read from VERSION.json in the project root
	return getVersionFromFile()
}

// getVersionFromFile reads the version from VERSION.json file (for development only)
func getVersionFromFile() string {
	// This is only used during development when Version = "dev"
	// Try to find VERSION.json in the project root
	var possiblePaths []string

	// 1. Try current working directory first
	if cwd, err := os.Getwd(); err == nil {
		possiblePaths = append(possiblePaths, filepath.Join(cwd, "VERSION.json"))
	}

	// 2. Try relative to current directory (for development)
	possiblePaths = append(possiblePaths, "VERSION.json")

	// 3. Try going up one directory (in case we're in a subdirectory)
	possiblePaths = append(possiblePaths, filepath.Join("..", "VERSION.json"))

	for _, versionFile := range possiblePaths {
		if data, err := ioutil.ReadFile(versionFile); err == nil {
			var versionData VersionFile
			if err := json.Unmarshal(data, &versionData); err == nil && versionData.Version != "" {
				return versionData.Version
			}
		}
	}

	// Fallback to build-time version
	return Version
}

// CheckForUpdates checks for newer releases by comparing local VERSION.json with remote VERSION.json
func CheckForUpdates(githubRepo string) UpdateInfo {
	current := GetBuildInfo()

	updateInfo := UpdateInfo{
		Available:      false,
		CurrentVersion: current.Version,
	}

	// Don't check for updates if we're in dev mode
	if current.Version == "dev" {
		updateInfo.Error = "Development version - update checking disabled"
		return updateInfo
	}

	// Fetch VERSION.json directly from GitHub repository (raw content)
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/refs/heads/master/VERSION.json", githubRepo)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		updateInfo.Error = fmt.Sprintf("Failed to check for updates: %v", err)
		return updateInfo
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		updateInfo.Error = fmt.Sprintf("Failed to fetch VERSION.json: HTTP %d", resp.StatusCode)
		return updateInfo
	}

	var remoteVersion VersionFile
	if err := json.NewDecoder(resp.Body).Decode(&remoteVersion); err != nil {
		updateInfo.Error = fmt.Sprintf("Failed to parse remote VERSION.json: %v", err)
		return updateInfo
	}

	if remoteVersion.Version == "" {
		updateInfo.Error = "Remote VERSION.json contains empty version"
		return updateInfo
	}

	updateInfo.LatestVersion = remoteVersion.Version
	updateInfo.ReleaseURL = fmt.Sprintf("https://github.com/%s/releases", githubRepo)
	updateInfo.ReleaseNotes = fmt.Sprintf("Version %s is available", remoteVersion.Version)

	// Compare the local VERSION.json version with the remote VERSION.json version
	updateInfo.Available = isNewerVersion(updateInfo.LatestVersion, current.Version)

	return updateInfo
}

// isNewerVersion compares two semantic version strings
// Returns true if latest is newer than current
func isNewerVersion(latest, current string) bool {
	// Clean up version strings (remove 'v' prefix if present)
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// If versions are the same, no update needed
	if latest == current {
		return false
	}

	// Split versions into parts
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Pad the shorter version with zeros
	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}

	// Compare each part numerically
	for i := 0; i < maxLen; i++ {
		latestNum := parseVersionPart(latestParts[i])
		currentNum := parseVersionPart(currentParts[i])

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
		// If equal, continue to next part
	}

	return false // Versions are equal
}

// parseVersionPart extracts numeric part from version component
func parseVersionPart(part string) int {
	// Extract numeric part only (ignore any suffix like "rc1", "beta", etc.)
	numStr := ""
	for _, char := range part {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	// Convert to int manually
	num := 0
	for _, digit := range numStr {
		if digit >= '0' && digit <= '9' {
			num = num*10 + int(digit-'0')
		}
	}
	return num
}

// FormatVersion returns a formatted version string for display
func FormatVersion() string {
	info := GetBuildInfo()
	if info.Version == "dev" {
		commit := strings.TrimPrefix(info.Commit, "unknown")
		if len(commit) > 8 {
			commit = commit[:8]
		} else if commit == "" {
			commit = "unknown"
		}
		return fmt.Sprintf("Development Build (%s)", commit)
	}
	return info.Version
}

// GetDetailedVersionInfo returns detailed version information for display
func GetDetailedVersionInfo() map[string]string {
	info := GetBuildInfo()

	details := map[string]string{
		"version":   info.Version,
		"commit":    info.Commit,
		"buildDate": info.Date,
		"goVersion": info.GoVersion,
	}
	// Format commit for display
	if info.Commit != "unknown" && len(info.Commit) > 8 {
		details["shortCommit"] = info.Commit[:8]
	} else {
		details["shortCommit"] = info.Commit
	}

	// Format build date for display
	if info.Date != "unknown" {
		if t, err := time.Parse(time.RFC3339, info.Date); err == nil {
			details["formattedDate"] = t.Format("January 2, 2006 at 15:04 MST")
		} else {
			details["formattedDate"] = info.Date
		}
	} else {
		details["formattedDate"] = "Unknown"
	}

	return details
}
