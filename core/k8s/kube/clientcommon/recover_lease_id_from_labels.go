package clientcommon

import (
	"errors"
	"fmt"
	"strconv"

	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
)

var (
	errMissingLabel      = errors.New("kube: missing label")
	errInvalidLabelValue = errors.New("kube: invalid label value")
)

func RecoverLeaseIDFromLabels(labels map[string]string) (mtypes.LeaseID, error) {
	dseqS, ok := labels[builder.SubnetNodeLeaseDSeqLabelName]
	if !ok {
		return mtypes.LeaseID{}, fmt.Errorf("%w: %q", errMissingLabel, builder.SubnetNodeLeaseDSeqLabelName)
	}
	gseqS, ok := labels[builder.SubnetNodeLeaseGSeqLabelName]
	if !ok {
		return mtypes.LeaseID{}, fmt.Errorf("%w: %q", errMissingLabel, builder.SubnetNodeLeaseGSeqLabelName)
	}
	oseqS, ok := labels[builder.SubnetNodeLeaseOSeqLabelName]
	if !ok {
		return mtypes.LeaseID{}, fmt.Errorf("%w: %q", errMissingLabel, builder.SubnetNodeLeaseOSeqLabelName)
	}
	owner, ok := labels[builder.SubnetNodeLeaseOwnerLabelName]
	if !ok {
		return mtypes.LeaseID{}, fmt.Errorf("%w: %q", errMissingLabel, builder.SubnetNodeLeaseOwnerLabelName)
	}

	provider, ok := labels[builder.SubnetNodeLeaseProviderLabelName]
	if !ok {
		return mtypes.LeaseID{}, fmt.Errorf("%w: %q", errMissingLabel, builder.SubnetNodeLeaseProviderLabelName)
	}

	dseq, err := strconv.ParseUint(dseqS, 10, 64)
	if err != nil {
		return mtypes.LeaseID{}, fmt.Errorf("%w: dseq %q not a uint", errInvalidLabelValue, dseqS)
	}

	gseq, err := strconv.ParseUint(gseqS, 10, 32)
	if err != nil {
		return mtypes.LeaseID{}, fmt.Errorf("%w: gseq %q not a uint", errInvalidLabelValue, gseqS)
	}

	oseq, err := strconv.ParseUint(oseqS, 10, 32)
	if err != nil {
		return mtypes.LeaseID{}, fmt.Errorf("%w: oesq %q not a uint", errInvalidLabelValue, oseqS)
	}

	return mtypes.LeaseID{
		Owner:    owner,
		DSeq:     dseq,
		GSeq:     uint32(gseq),
		OSeq:     uint32(oseq),
		Provider: provider,
	}, nil
}
