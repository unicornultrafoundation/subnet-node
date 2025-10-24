package dynamic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

type contextKey string

const (
	nonceContextKey     contextKey = "nonce"
	signatureContextKey contextKey = "signature"
)

type HTTPClient struct {
	ip        string
	pubkey    string
	serverURL string
	host      host.Host
	client    *http.Client
}

var _ DynamicIPClient = (*HTTPClient)(nil)

var log = logrus.WithField("service", "vpn-dhcp-http-client")

func NewHTTPClient(cfg *config.C, host host.Host) (DynamicIPClient, error) {
	pubkey := host.Peerstore().PrivKey(host.ID()).GetPublic()
	pubkeyBytes, err := crypto.MarshalPublicKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("unmarshal public key: %w", err)
	}
	pubkeyString := base64.StdEncoding.EncodeToString(pubkeyBytes)

	serverURL := cfg.GetString("vpn.dhcp_url", "")

	if serverURL == "" {
		return nil, fmt.Errorf("vpn.dhcp_url is not configured")
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	return &HTTPClient{serverURL: serverURL, pubkey: pubkeyString, client: client, host: host}, nil
}

func (c *HTTPClient) GetIP(ctx context.Context) (string, error) {
	return c.ip, nil
}

func (c *HTTPClient) SetIP(ctx context.Context, ip string) error {
	c.ip = ip
	return nil
}

func (c *HTTPClient) GetLeaseByPeerID(ctx context.Context, peerID string) (Lease, error) {
	if peerID == "" {
		return Lease{}, fmt.Errorf("peerID is required")
	}

	var lease Lease
	err := c.doRequest(ctx, "GET", "/lease/peer-id/"+peerID, nil, &lease)
	if err != nil {
		return Lease{}, err
	}

	return lease, nil
}

func (c *HTTPClient) GetLeaseByIP(ctx context.Context, ip string) (Lease, error) {
	if ip == "" {
		return Lease{}, fmt.Errorf("ip is required")
	}

	// Convert the IP to a token ID
	tokenID := utils.ConvertVirtualIPToNumber(ip)
	if tokenID == 0 {
		return Lease{}, fmt.Errorf("invalid IP: %s", ip)
	}

	var lease Lease
	err := c.doRequest(ctx, "GET", "/lease/token-id/"+strconv.FormatInt(int64(tokenID), 10), nil, &lease)
	if err != nil {
		return Lease{}, err
	}

	return lease, nil
}

func (c *HTTPClient) RequestIP(ctx context.Context) (Lease, error) {
	var lease Lease
	err := c.doAuthRequest(ctx, "POST", "/allocate-ip", nil, &lease)
	if err != nil {
		return Lease{}, err
	}

	return lease, nil
}

func (c *HTTPClient) RenewIP(ctx context.Context) (Lease, error) {
	tokenID := utils.ConvertVirtualIPToNumber(c.ip)
	if tokenID == 0 {
		return Lease{}, fmt.Errorf("invalid IP: %s", c.ip)
	}
	path := fmt.Sprintf("/renew-lease?tokenID=%d", tokenID)

	var lease Lease
	err := c.doAuthRequest(ctx, "POST", path, nil, &lease)
	if err != nil {
		return Lease{}, err
	}

	return lease, nil
}

func (c *HTTPClient) ReleaseIP(ctx context.Context) error {
	tokenID := utils.ConvertVirtualIPToNumber(c.ip)
	if tokenID == 0 {
		return fmt.Errorf("invalid IP: %s", c.ip)
	}
	path := fmt.Sprintf("/release-lease?tokenID=%d", tokenID)

	err := c.doAuthRequest(ctx, "POST", path, nil, nil)
	if err != nil {
		return fmt.Errorf("release ip: %w", err)
	}

	return err
}

func (c *HTTPClient) doAuthRequest(
	ctx context.Context,
	method, path string,
	reqBody interface{},
	respBody interface{},
) error {
	var authResponse AuthResponse
	// Request nonce from server
	err := c.doRequest(ctx, "POST", "/request-auth", nil, &authResponse)
	if err != nil {
		return fmt.Errorf("request auth: %w", err)
	}

	hashNonce := sha256.Sum256([]byte(authResponse.Nonce))

	// Sign the nonce
	signature, err := c.host.Peerstore().PrivKey(c.host.ID()).Sign(hashNonce[:])
	if err != nil {
		return fmt.Errorf("sign nonce: %w", err)
	}

	// Attach nonce and signature to ctx
	ctx = context.WithValue(ctx, nonceContextKey, authResponse.Nonce)
	ctx = context.WithValue(ctx, signatureContextKey, signature)

	return c.doRequest(ctx, method, path, reqBody, respBody)
}

func (c *HTTPClient) doRequest(
	ctx context.Context,
	method, path string,
	reqBody interface{},
	respBody interface{},
) error {
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.serverURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PubKey", c.pubkey)
	if signature, ok := ctx.Value(signatureContextKey).([]byte); ok {
		req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(signature))
	}
	if nonce, ok := ctx.Value(nonceContextKey).(string); ok {
		req.Header.Set("X-Nonce", nonce)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("http %s %s failed: %s (%s)", method, path, res.Status, string(msg))
	}

	if respBody != nil {
		// Read entire body to allow flexible decoding (supports optional {"data": ...} envelope)
		responseBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}
		if len(responseBytes) == 0 {
			return nil
		}

		// Try to unwrap {"data": ...} envelope first
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(responseBytes, &envelope); err == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
			if err := json.Unmarshal(envelope.Data, respBody); err != nil {
				return fmt.Errorf("decode enveloped response: %w", err)
			}
			return nil
		}

		// Fallback: decode the response directly into the expected body
		if err := json.Unmarshal(responseBytes, respBody); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
