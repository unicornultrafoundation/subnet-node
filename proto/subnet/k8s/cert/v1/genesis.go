package v1

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
)

type GenesisCertificates []GenesisCertificate

func (obj GenesisCertificates) Contains(cert GenesisCertificate) bool {
	for _, c := range obj {
		if c.Owner == cert.Owner {
			return true
		}

		if bytes.Equal(c.Certificate.Cert, cert.Certificate.Cert) {
			return true
		}
	}

	return false
}

func (m GenesisCertificate) Validate() error {
	if !common.IsHexAddress(m.Owner) {
		return ErrInvalidAddress
	}
	owner := common.HexToAddress(m.Owner)
	if err := m.Certificate.Validate(owner); err != nil {
		return err
	}

	return nil
}

func (m *GenesisState) Validate() error {
	for _, cert := range m.Certificates {
		if err := cert.Validate(); err != nil {
			return err
		}
	}
	return nil
}
