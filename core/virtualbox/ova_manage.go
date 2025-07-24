package virtualbox

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// FIX the link later
var defaultOVAURLs = map[string]string{
	"Ubuntu_64":    "https://www.sendgb.com/src/download_one.php?uploadId=QbC28risRfG&sc=ababb9417e3bb2e7617724b68c881df6&file=template_sample_Ubuntu_64.ova&private_id=",
	"Ubuntu_ARM64": "https://www.sendgb.com/src/download_one.php?uploadId=AGHUcYp37x2&sc=a4e01fb96cc851b6ef31eaf07807ec0c&file=template_sample_Ubuntu_ARM64.ova&private_id=",
}

// getOVAURLForOSType returns the OVA URL for the given OS type from the hardcoded map
func getOVAURLForOSType(osType string) (string, error) {
	url, ok := defaultOVAURLs[osType]
	if !ok {
		return "", fmt.Errorf("no OVA for OS type %s", osType)
	}
	return url, nil
}

// DefaultOVAURLs returns the hardcoded OVA URLs map
func DefaultOVAURLs() map[string]string {
	return defaultOVAURLs
}

// DownloadFile downloads a file from the given URL to the specified local path
func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// HTTPGetWithProgress wraps http.Get for progress reporting
func HTTPGetWithProgress(url string) (*http.Response, error) {
	return http.Get(url)
}

// CopyWithProgress copies from src to dst and prints progress to stdout
func CopyWithProgress(dst *os.File, src io.Reader, total int64) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	var lastPrint time.Time
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
			if time.Since(lastPrint) > 1*time.Second || written == total {
				if total > 0 {
					fmt.Printf("  Progress: %.1f%% (%d/%d MB)\n", float64(written)*100/float64(total), written/1024/1024, total/1024/1024)
				} else {
					fmt.Printf("  Progress: %d bytes\n", written)
				}
				lastPrint = time.Now()
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return written, er
		}
	}
	return written, nil
}
