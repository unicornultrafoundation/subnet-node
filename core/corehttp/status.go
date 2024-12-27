package corehttp

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/core"
)

func StatusOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			status := map[string]interface{}{
				"online": n.IsOnline,
				"id":     n.Identity,
				// Add more status fields as needed
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(status)
		})
		return mux, nil
	}
}
