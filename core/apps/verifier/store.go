package verifier

import (
	"context"
	"fmt"

	proto "github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-datastore"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

func (v *Verifier) logUsageReportToDB(report *pvtypes.UsageReport) error {
	// Convert UsageReport to a format suitable for the datastore
	data, err := proto.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal usage report: %v", err)
	}

	// Store the usage report in the datastore
	key := datastore.NewKey(fmt.Sprintf("/usage_reports/%d/%s/%d", report.AppId, report.PeerId, report.Timestamp))
	err = v.ds.Put(context.Background(), key, data)
	if err != nil {
		return fmt.Errorf("failed to store usage report in datastore: %v", err)
	}

	// Retrieve existing usage info
	usageInfo, err := v.getUsageInfoFromDB(report.AppId, report.PeerId)
	if err != nil {
		usageInfo = &pvtypes.UsageInfo{}
	}

	// Update usage info
	usageInfo.PreviousUsageReport = report

	// Store updated usage info
	err = v.storeUsageInfoToDB(report.AppId, report.PeerId, usageInfo)
	if err != nil {
		return fmt.Errorf("failed to store usage info in datastore: %v", err)
	}

	return nil
}

func (v *Verifier) getUsageInfoFromDB(appId int64, peerId string) (*pvtypes.UsageInfo, error) {
	// Retrieve the usage info from the datastore
	key := datastore.NewKey(fmt.Sprintf("/usage_info/%d/%s", appId, peerId))
	data, err := v.ds.Get(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve usage info from datastore: %v", err)
	}

	usageInfo := &pvtypes.UsageInfo{}
	err = proto.Unmarshal(data, usageInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage info: %v", err)
	}

	return usageInfo, nil
}

func (v *Verifier) storeUsageInfoToDB(appId int64, peerId string, usageInfo *pvtypes.UsageInfo) error {
	// Convert UsageInfo to a format suitable for the datastore
	data, err := proto.Marshal(usageInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal usage info: %v", err)
	}

	// Store the usage info in the datastore
	key := datastore.NewKey(fmt.Sprintf("/usage_info/%d/%s", appId, peerId))
	err = v.ds.Put(context.Background(), key, data)
	if err != nil {
		return fmt.Errorf("failed to store usage info in datastore: %v", err)
	}

	return nil
}

func (v *Verifier) getPreviousTimestampFromDB(appId int64, peerId string) (int64, error) {
	usageInfo, err := v.getUsageInfoFromDB(appId, peerId)
	if err != nil {
		return 0, err
	}
	return usageInfo.PreviousUsageReport.Timestamp, nil
}

func (v *Verifier) markPeerAsGood(appId int64, peerId string) error {
	// Mark the peer as good in the datastore
	key := datastore.NewKey(fmt.Sprintf("/good_peers/%d/%s", appId, peerId))
	err := v.ds.Put(context.Background(), key, []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark peer as good in datastore: %v", err)
	}
	return nil
}

func (v *Verifier) saveSignedUsages(w datastore.Write, signedUsage []*pvtypes.SignedUsage) error {
	for _, usage := range signedUsage {
		// Convert SignedUsage to a format suitable for the datastore
		data, err := proto.Marshal(usage)
		if err != nil {
			return fmt.Errorf("failed to marshal signed usage: %v", err)
		}

		// Store the signed usage in the datastore
		key := datastore.NewKey(fmt.Sprintf("/signed_usage/%d/%s/%d", usage.AppId, usage.PeerId, usage.Timestamp))
		err = w.Put(context.Background(), key, data)
		if err != nil {
			return fmt.Errorf("failed to store signed usage in datastore: %v", err)
		}

	}
	return nil
}
