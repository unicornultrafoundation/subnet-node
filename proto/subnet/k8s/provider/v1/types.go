package v1

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
)

// String implements the Stringer interface for a Provider object.
func (p Provider) String() string {
	res := fmt.Sprintf(`Deployment
	Owner:   %s
	HostURI: %s
	Attributes: %v
	`, p.Owner, p.HostURI, p.Attributes)

	if !p.Info.IsEmpty() {
		res += fmt.Sprintf("Info: %v\n", p.Info)
	}
	return res
}

// Providers is the collection of Provider
type Providers []Provider

// String implements the Stringer interface for a Providers object.
func (obj Providers) String() string {
	var buf bytes.Buffer

	const sep = "\n\n"

	for _, p := range obj {
		buf.WriteString(p.String())
		buf.WriteString(sep)
	}

	if len(obj) > 0 {
		buf.Truncate(buf.Len() - len(sep))
	}

	return buf.String()
}

// Address implements provider and returns owner of provider
func (p *Provider) Address() common.Address {
	if !common.IsHexAddress(p.Owner) {
		panic(ErrInvalidAddress)
	}
	return common.HexToAddress(p.Owner)
}

func (m ProviderInfo) IsEmpty() bool {
	return m.EMail == "" && m.Website == ""
}

func (m ProviderInfo) Validate() error {
	if m.Website != "" {
		if _, err := url.Parse(m.Website); err != nil {
			return ErrInvalidInfoWebsite
		}
	}
	return nil
}
