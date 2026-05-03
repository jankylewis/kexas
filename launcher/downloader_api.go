package launcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

// getChromiumDownloadURL returns the download URL for the current platform.
// Fetches from Chrome for Testing JSON API.
func getChromiumDownloadURL(version string) (string, error) {
	var apiResp *ChromeForTestingResponse
	var err error
	apiResp, err = fetchChromeForTestingResponse()
	if err != nil {
		return "", err
	}

	var channelData ChromeChannelData
	var ok bool
	channelData, ok = apiResp.Channels[ChromeReleaseChannel]
	if !ok {
		return "", fmt.Errorf("launcher: %s channel not found", ChromeReleaseChannel)
	}

	var chromeDownloads []ChromeDownload
	chromeDownloads, ok = channelData.Downloads["chrome"]
	if !ok {
		return "", fmt.Errorf("launcher: chrome downloads not found")
	}

	var platform string
	platform, err = currentPlatformID()
	if err != nil {
		return "", err
	}

	for _, download := range chromeDownloads {
		if download.Platform == platform {
			return download.URL, nil
		}
	}
	return "", fmt.Errorf("launcher: no download found for platform %s", platform)
}

// fetchChromeForTestingResponse GETs the CfT API and decodes the JSON response.
func fetchChromeForTestingResponse() (*ChromeForTestingResponse, error) {
	var resp *http.Response
	var err error
	resp, err = http.Get(ChromeForTestingAPIURL)
	if err != nil {
		return nil, fmt.Errorf("launcher: failed to fetch download URLs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launcher: API returned status %s", resp.Status)
	}

	var apiResp ChromeForTestingResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return nil, fmt.Errorf("launcher: failed to parse API response: %w", err)
	}
	return &apiResp, nil
}

// currentPlatformID returns the Chrome-for-Testing platform identifier
// for the current GOOS/GOARCH (e.g. "mac-arm64", "linux64", "win64").
func currentPlatformID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "mac-arm64", nil
		}
		return "mac-x64", nil
	case "linux":
		return "linux64", nil
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "win64", nil
		}
		return "win32", nil
	default:
		return "", fmt.Errorf("launcher: unsupported platform: %s", runtime.GOOS)
	}
}
