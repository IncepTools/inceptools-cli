package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"incepttools/src/ui"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// GithubRelease represents the structure of a GitHub release from the API
type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func HandleUpdate(currentVersion string) {
	ui.Info("Current version: %s", currentVersion)
	ui.Info("🔍 Fetching releases from GitHub...")

	resp, err := http.Get("https://api.github.com/repos/IncepTools/inceptools-cli/releases")
	if err != nil {
		ui.Error("Failed to fetch releases: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.Error("Failed to fetch releases: %s", resp.Status)
		return
	}

	var releases []GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		ui.Error("Failed to parse releases: %v", err)
		return
	}

	if len(releases) == 0 {
		ui.Info("No releases found.")
		return
	}

	fmt.Println("\nAvailable versions:")
	for i, rel := range releases {
		fmt.Printf("[%d] %s\n", i+1, rel.TagName)
		if i >= 4 { // Show only last 5 releases
			break
		}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nSelect version to install (default: 1): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx := 0
	if input != "" {
		idx, err = strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(releases) {
			ui.Error("Invalid selection")
			return
		}
		idx-- // 0-based
	}

	targetRelease := releases[idx]
	ui.Info("🚀 Updating to %s...", targetRelease.TagName)

	// Find the correct asset for current OS/Arch
	suffix := fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	var downloadURL string
	for _, asset := range targetRelease.Assets {
		if strings.HasSuffix(asset.Name, suffix) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		ui.Error("Could not find a compatible binary for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, targetRelease.TagName)
		return
	}

	updateBinary(downloadURL)
}

func updateBinary(url string) {
	ui.Info("📥 Downloading...")
	resp, err := http.Get(url)
	if err != nil {
		ui.Error("Download failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.Error("Download failed with status: %s", resp.Status)
		return
	}

	executablePath, err := os.Executable()
	if err != nil {
		ui.Error("Could not determine executable path: %v", err)
		return
	}

	tmpPath := executablePath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		ui.Error("Failed to create temporary file: %v", err)
		return
	}
	defer os.Remove(tmpPath)

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		ui.Error("Failed to save new binary: %v", err)
		return
	}

	oldPath := executablePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(executablePath, oldPath); err != nil {
		ui.Error("Failed to move current binary: %v", err)
		return
	}

	if err := os.Rename(tmpPath, executablePath); err != nil {
		os.Rename(oldPath, executablePath)
		ui.Error("Failed to replace binary: %v", err)
		return
	}

	os.Remove(oldPath)
	ui.Success("Update successful! Please restart inceptools.")
}
