package utils

import "encoding/json"

// SortJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error.
//
// Deprecated: SortJSON was used for GetSignbytes, this is now automatic with amino signing
func SortJSON(toSortJSON []byte) ([]byte, error) {
	var c any
	err := json.Unmarshal(toSortJSON, &c)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return js, nil
}
