package manifest

import (
	"encoding/base32"
	"regexp"
	"strings"

	"github.com/google/uuid"

	manitypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

func AllHostnamesOfManifestGroup(mgroup manitypes.Group) []string {
	allHostnames := make([]string, 0)
	for _, service := range mgroup.Services {
		for _, expose := range service.Expose {
			allHostnames = append(allHostnames, expose.Hosts...)
		}
	}

	return allHostnames
}

func IngressHost(lid mtypes.LeaseID, svcName string) string {
	uid := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(lid.String()+svcName))
	// MarshalBinary always returns nil
	data, _ := uid.MarshalBinary()
	return strings.ToLower(base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(data))
}

func SanitizeSubdomain(name string) string {
	// Convert to lowercase (DNS is case-insensitive)
	name = strings.ToLower(name)

	// Keep only letters, digits, and hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "")

	// Trim leading and trailing hyphens
	name = strings.Trim(name, "-")

	// Limit to 55 characters (max subdomain label length - 8 for the UID)
	if len(name) > 55 {
		name = name[:55]
	}

	return name
}
