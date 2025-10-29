package v1

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "market"

	// StoreKey is the store key string for market
	StoreKey = ModuleName

	// RouterKey is the message route for market
	RouterKey = ModuleName
)

func OrderPrefix() []byte {
	return []byte{0x01, 0x00}
}

func BidPrefix() []byte {
	return []byte{0x02, 0x00}
}

func LeasePrefix() []byte {
	return []byte{0x03, 0x00}
}

func SecondaryLeasePrefix() []byte {
	return []byte{0x03, 0x01}
}

type LeaseIDKey struct {
	Owner    string
	DSeq     uint64
	GSeq     uint32
	OSeq     uint32
	Provider string
}

func (k LeaseIDKey) ToLeaseID() LeaseID {
	return LeaseID{
		Owner:    k.Owner,
		DSeq:     k.DSeq,
		GSeq:     k.GSeq,
		OSeq:     k.OSeq,
		Provider: k.Provider,
	}
}

func LeaseIDToKey(id LeaseID) LeaseIDKey {
	return LeaseIDKey{
		Owner:    id.Owner,
		DSeq:     id.DSeq,
		GSeq:     id.GSeq,
		OSeq:     id.OSeq,
		Provider: id.Provider,
	}
}
