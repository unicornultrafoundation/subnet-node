package manifest

import (
	"encoding/base32"
	"strings"

	"github.com/google/uuid"
	maniv2beta1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

func AllHostnamesOfManifestGroup(mgroup maniv2beta1.Group) []string {
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
