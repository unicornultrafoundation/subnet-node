package storage

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type ImageRequest struct {
	OS      string `json:"os"`
	Version string `json:"version"`
}

var ubuntuVersionToCodename = map[string]string{
	"24.04": "noble",
	"22.04": "jammy",
	"20.04": "focal",
	"18.04": "bionic",
	"16.04": "xenial",
	"14.04": "trusty",
	"23.10": "mantic",
	"23.04": "lunar",
	"22.10": "kinetic",
	"21.10": "impish",
	"21.04": "hirsute",
	"20.10": "groovy",
	"19.10": "eoan",
	"19.04": "disco",
	"18.10": "cosmic",
}

func (is *ImageStorage) GenerateImageURL(req ImageRequest) (string, error) {

	if strings.ToLower(req.OS) != "ubuntu" {
		return "", fmt.Errorf("unsupported OS: %s", req.OS)
	}

	codename, exists := ubuntuVersionToCodename[req.Version]
	if !exists {
		return "", fmt.Errorf("unsupported Ubuntu version: %s", req.Version)
	}

	arch := runtime.GOARCH
	if arch == "" {
		arch = "amd64"
	}

	latestDate, err := is.getLatestDate(codename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://cloud-images.ubuntu.com/%s/%s/%s-server-cloudimg-%s.img",
		codename, latestDate, codename, arch), nil
}

// getLatestDate fetches the directory listing and finds the latest date
func (is *ImageStorage) getLatestDate(codename string) (string, error) {
	url := fmt.Sprintf("https://cloud-images.ubuntu.com/%s/", codename)

	resp, err := is.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse HTML to find date directories
	dates := is.parseDateDirectories(string(body))
	if len(dates) == 0 {
		return "", fmt.Errorf("no date directories found")
	}

	// Sort dates and return the latest
	sort.Strings(dates)
	return dates[len(dates)-1], nil
}

// parseDateDirectories extracts date directories from HTML listing
func (is *ImageStorage) parseDateDirectories(html string) []string {
	// Match date patterns like "20250626/" in href attributes
	datePattern := regexp.MustCompile(`href="(\d{8})/"`)
	matches := datePattern.FindAllStringSubmatch(html, -1)

	var dates []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			date := match[1]
			if !seen[date] {
				dates = append(dates, date)
				seen[date] = true
			}
		}
	}

	return dates
}

// GenerateImageURLWithFallback generates URL with fallback to a known good date
func (is *ImageStorage) GenerateImageURLWithFallback(req ImageRequest) string {
	url, err := is.GenerateImageURL(req)
	if err != nil {
		// Fallback to current date format if API fails
		codename, exists := ubuntuVersionToCodename[req.Version]
		if !exists {
			return ""
		}

		arch := runtime.GOARCH
		url = fmt.Sprintf("https://cloud-images.ubuntu.com/%s/%s/%s-server-cloudimg-%s.img",
			codename, time.Now().Format("20060102"), codename, arch)
	}
	return url
}
