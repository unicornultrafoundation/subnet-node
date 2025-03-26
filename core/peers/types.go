package peers

type PeerMultiAddress struct {
	PeerID       string   `json:"peer_id"`
	MultiAddresses []string `json:"multi_addresses"`
}
