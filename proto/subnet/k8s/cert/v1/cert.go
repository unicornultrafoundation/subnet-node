package v1

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func ParseAndValidateCertificate(owner common.Address, crt, pub []byte) (*x509.Certificate, error) {
	blk, rest := pem.Decode(pub)
	if blk == nil || len(rest) > 0 {
		return nil, ErrInvalidPubkeyValue
	}

	if blk.Type != PemBlkTypeECPublicKey {
		return nil, errors.Wrap(ErrInvalidPubkeyValue, "invalid pem block type")
	}

	blk, rest = pem.Decode(crt)
	if blk == nil || len(rest) > 0 {
		return nil, ErrInvalidCertificateValue
	}

	if blk.Type != PemBlkTypeCertificate {
		return nil, errors.Wrap(ErrInvalidCertificateValue, "invalid pem block type")
	}

	cert, err := x509.ParseCertificate(blk.Bytes)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(cert.Subject.CommonName) {
		return nil, errors.Wrap(ErrInvalidCertificateValue, "invalid CommonName")
	}

	cowner := common.HexToAddress(cert.Subject.CommonName)
	if owner.Cmp(cowner) != 0 {
		return nil, errors.Wrap(ErrInvalidCertificateValue, "CommonName does not match owner")
	}

	return cert, nil
}

func (m *CertificateID) String() string {
	return fmt.Sprintf("%s/%s", m.Owner, m.Serial)
}

func (m *CertificateID) Equals(val CertificateID) bool {
	return (m.Owner == val.Owner) && (m.Serial == val.Serial)
}

func (m Certificate) Validate(owner common.Address) error {
	if val, exists := Certificate_State_name[int32(m.State)]; !exists || val == "invalid" {
		return ErrInvalidState
	}

	_, err := ParseAndValidateCertificate(owner, m.Cert, m.Pubkey)
	if err != nil {
		return err
	}

	return nil
}

func (m Certificate) IsState(state Certificate_State) bool {
	return m.State == state
}
