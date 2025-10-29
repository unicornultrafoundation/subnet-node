package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/remotecommand"

	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	manifest "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

const (
	schemeWSS   = "wss"
	schemeHTTPS = "https"
)

const (
	contentTypeJSON = "application/json; charset=UTF-8"

	// PingWait Time allowed writing the file to the client.
	PingWait = 15 * time.Second

	// PongWait Time allowed reading the next pong message from the client.
	PongWait = 15 * time.Second

	// PingPeriod Send pings to a client with this period. Must be less than pongWait.
	PingPeriod = 10 * time.Second
)

var (
	ErrNotInitialized = errors.New("rest: not initialized")
)

type ReqClient interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
	Do(*http.Request) (*http.Response, error)
}

// Client defines the methods available for connecting to the gateway server.
type Client interface {
	Status(ctx context.Context) (*ProviderStatus, error)
	Validate(ctx context.Context, gspec dtypes.GroupSpec) (ValidateGroupSpecResult, error)
	SubmitManifest(ctx context.Context, dseq uint64, mani manifest.Manifest) error
	GetManifest(ctx context.Context, id mtypes.LeaseID) (manifest.Manifest, error)
	LeaseStatus(ctx context.Context, id mtypes.LeaseID) (LeaseStatus, error)
	LeaseEvents(ctx context.Context, id mtypes.LeaseID, services string, follow bool) (*LeaseKubeEvents, error)
	LeaseLogs(ctx context.Context, id mtypes.LeaseID, services string, follow bool, tailLines int64) (*ServiceLogs, error)
	ServiceStatus(ctx context.Context, id mtypes.LeaseID, service string) (*ServiceStatus, error)
	LeaseShell(ctx context.Context, id mtypes.LeaseID, service string, podIndex uint, cmd []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		tty bool,
		tsq <-chan remotecommand.TerminalSize) error
	MigrateHostnames(ctx context.Context, hostnames []string, dseq uint64, gseq uint32) error
	MigrateEndpoints(ctx context.Context, endpoints []string, dseq uint64, gseq uint32) error
}

type client struct {
	ctx    context.Context
	host   *url.URL
	tlsCfg *tls.Config
}

type reqClient struct {
	ctx      context.Context
	host     *url.URL
	hclient  *http.Client
	wsclient *websocket.Dialer
}

// NewClient creates and returns a new Client instance for interacting with the Subnet provider.
//
// It takes a context.Context for managing the lifecycle of operations, a QueryClient for making
// provider queries, the provider's address, and optional ClientOption functions for customizing
// the client configuration.
//
// The following options can be provided.
//   - WithAuthCerts: Configure TLS certificates for secure communication
//   - WithAuthJWTSigner: Set a JWT signer for authentication
//   - WithAuthToken: Provide an authentication token
//
// Note, auth have the following priority: WithAuthCerts > WithAuthJWTSigner > WithAuthToken
//
// The function will:
// 1. Apply any provided ClientOptions
// 2. Query the provider's host URI using the QueryClient
// 3. Set up TLS configuration with system certificates
// 4. Configure client authentication using either provided certificates or JWT signing
//
// Returns an error if:
// - Any ClientOption fails to apply
// - The provider query fails
// - The host URI is invalid
// - System certificates cannot be loaded
func NewClient(ctx context.Context) (Client, error) {
	return &client{}, nil
}

func (c *reqClient) Do(req *http.Request) (*http.Response, error) {
	return c.hclient.Do(req)
}

func (c *reqClient) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	return c.wsclient.DialContext(ctx, urlStr, requestHeader)
}

type ClientResponseError struct {
	Status  int
	Message string
}

func (err ClientResponseError) Error() string {
	return fmt.Sprintf("remote server returned %d", err.Status)
}

func (err ClientResponseError) ClientError() string {
	return fmt.Sprintf("Remote Server returned %d\n%s", err.Status, err.Message)
}

func (c *client) Status(ctx context.Context) (*ProviderStatus, error) {
	uri, err := MakeURI(c.host, StatusPath())
	if err != nil {
		return nil, err
	}
	var obj ProviderStatus

	if err := c.getStatus(ctx, uri, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func (c *client) Validate(ctx context.Context, gspec dtypes.GroupSpec) (ValidateGroupSpecResult, error) {
	return ValidateGroupSpecResult{}, nil
}

func (c *client) SubmitManifest(ctx context.Context, dseq uint64, mani manifest.Manifest) error {

	return nil
}

func (c *client) GetManifest(ctx context.Context, lid mtypes.LeaseID) (manifest.Manifest, error) {
	return manifest.Manifest{}, nil
}

func (c *client) MigrateEndpoints(ctx context.Context, endpoints []string, dseq uint64, gseq uint32) error {
	return nil
}

func (c *client) MigrateHostnames(ctx context.Context, hostnames []string, dseq uint64, gseq uint32) error {
	return nil
}

func (c *client) LeaseStatus(ctx context.Context, id mtypes.LeaseID) (LeaseStatus, error) {
	uri, err := MakeURI(c.host, LeaseStatusPath(id))
	if err != nil {
		return LeaseStatus{}, err
	}

	var obj LeaseStatus
	if err := c.getStatus(ctx, uri, &obj); err != nil {
		return LeaseStatus{}, err
	}

	return obj, nil
}

func (c *client) LeaseEvents(ctx context.Context, id mtypes.LeaseID, _ string, follow bool) (*LeaseKubeEvents, error) {
	return &LeaseKubeEvents{}, nil
}

func (c *client) ServiceStatus(ctx context.Context, id mtypes.LeaseID, service string) (*ServiceStatus, error) {
	uri, err := MakeURI(c.host, ServiceStatusPath(id, service))
	if err != nil {
		return nil, err
	}

	var obj ServiceStatus
	if err := c.getStatus(ctx, uri, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func (c *client) getStatus(ctx context.Context, uri string, obj interface{}) error {
	return nil
}

func createClientResponseErrorIfNotOK(resp *http.Response, responseBuf *bytes.Buffer) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return ClientResponseError{
		Status:  resp.StatusCode,
		Message: responseBuf.String(),
	}
}

// MakeURI
// for client queries path must not include owner id
func MakeURI(uri *url.URL, path string) (string, error) {
	endpoint, err := url.Parse(uri.String() + "/" + path)
	if err != nil {
		return "", err
	}

	return endpoint.String(), nil
}

func (c *client) LeaseLogs(ctx context.Context,
	id mtypes.LeaseID,
	services string,
	follow bool,
	_ int64) (*ServiceLogs, error) {

	return &ServiceLogs{}, nil
}

// parseCloseMessage extract close reason from websocket close message
// "websocket: [error code]: [client reason]"
func parseCloseMessage(msg string) string {
	errmsg := strings.SplitN(msg, ": ", 3)
	if len(errmsg) == 3 {
		return errmsg[2]
	}

	return ""
}
