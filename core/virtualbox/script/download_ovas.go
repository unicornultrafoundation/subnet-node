package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
)

func main() {
	// Optionally: accept OS type as argument
	var osTypes []string
	if len(os.Args) > 1 {
		osTypes = os.Args[1:]
	} else {
		for osType := range virtualbox.DefaultOVAURLs() {
			osTypes = append(osTypes, osType)
		}
	}

	templatesDir, _ := fsutil.ExpandHome("~/VirtualBox VMs/Templates")
	if err := fsutil.DirWritable(templatesDir); err != nil {
		fmt.Printf("Failed to ensure Templates dir: %v\n", err)
		os.Exit(1)
	}

	for _, osType := range osTypes {
		url, ok := virtualbox.DefaultOVAURLs()[osType]
		if !ok {
			fmt.Printf("No URL for OS type: %s\n", osType)
			continue
		}
		ovaPath := filepath.Join(templatesDir, fmt.Sprintf("template_sample_%s.ova", osType))
		if fsutil.FileExists(ovaPath) {
			fmt.Printf("[SKIP] %s already exists at %s\n", osType, ovaPath)
			continue
		}
		fmt.Printf("[START] Downloading %s to %s\n", osType, ovaPath)
		start := time.Now()
		err := DownloadFileWithProgress(ovaPath, url)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("[FAIL] %s: %v\n", osType, err)
		} else {
			fmt.Printf("[DONE] %s in %s\n", osType, dur.Round(time.Second))
		}
	}
}

// DownloadFileWithProgress downloads a file and prints progress to stdout
func DownloadFileWithProgress(filepath string, url string) error {
	resp, err := virtualbox.HTTPGetWithProgress(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	written, err := virtualbox.CopyWithProgress(out, resp.Body, resp.ContentLength)
	if err == nil {
		fmt.Printf("  Downloaded %d bytes\n", written)
	}
	return err
}
