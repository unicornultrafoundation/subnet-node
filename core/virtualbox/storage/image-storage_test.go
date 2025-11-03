package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expandHome(p string) string {
	if p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return p
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
		return p
	}
	return p
}

func TestImageStorage_GetOrCreateVDIa(t *testing.T) {
	vmRoot := expandHome("~/VirtualBox VMs")
	imagesRoot := filepath.Join(vmRoot, "Images")
	imageStorage := NewImageStorage(vmRoot, imagesRoot)
	imageStorage.GetOrCreateVDI(context.Background(), "24.04", "arm64", "test-vm")
	t.Log("test-vm", "24.04", "arm64")
}
