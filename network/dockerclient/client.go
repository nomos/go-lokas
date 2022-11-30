package dockerclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	userAgent = "go-lokas-dockercli"

	unixProtocol = "unix"
)

var (
	// ErrInvalidEndpoint is returned when the endpoint is not a valid URL.
	ErrInvalidEndpoint = errors.New("invalid endpoint")
	// ErrConnectionRefused is returned when the client cannot connect to the given endpoint.
	ErrConnectionRefused = errors.New("cannot connect to Docker endpoint")
)

type Client struct {
	endpointURL *url.URL

	HTTPClient *http.Client
	TLSConfig  *tls.Config
}

func NewClient(endpoint string) (*Client, error) {
	//TODO: Select engine api version
	client, err := NewVersionedClient(endpoint)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewVersionedClient(endpoint string) (*Client, error) {
	u, err := parseEndPoint(endpoint)
	if err != nil {
		return nil, err
	}
	c := &Client{
		endpointURL: u,
		HTTPClient:  defaultClient(),
	}
	return c, nil
}

// defaultClient returns a new http.Client
func defaultClient() *http.Client {
	return &http.Client{}
}

// NewTLSClient returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates, using a specific remote API version.
// TODO: 🚧 Under construction...
func NewTLSClient(endpoint string, dockerCertPath string) (*Client, error) {
	caFile := filepath.Join(dockerCertPath, "ca.pem")
	certFile := filepath.Join(dockerCertPath, "cert.pem")
	keyFile := filepath.Join(dockerCertPath, "key.pem")

	if _, err := os.Stat(caFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return nil, err
	}

	u, err := parseEndPoint(endpoint)
	if err != nil {
		return nil, err
	}

	if err := os.Setenv("DOCKER_CERT_PATH", dockerCertPath); err != nil {
		return nil, err
	}
	if err := os.Setenv("DOCKER_HOST", endpoint); err != nil {
		return nil, err
	}

	ret := &Client{
		endpointURL: u,
	}
	return ret, nil
}

func parseEndPoint(endpoint string) (*url.URL, error) {
	if endpoint != "" && !strings.Contains(endpoint, "://") {
		endpoint = "tcp://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}

	switch u.Scheme {
	case "http", "https", "tcp":
		_, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			var e *net.AddrError
			if errors.As(err, &e) {
				if e.Err == "missing port in address" {
					return u, nil
				}
			}
			return nil, ErrInvalidEndpoint
		}
		number, err := strconv.ParseInt(port, 10, 64)
		if err == nil && number > 0 && number < 65536 {
			if u.Scheme == "tcp" {
				u.Scheme = "http"
			}
			return u, nil
		}
		return nil, ErrInvalidEndpoint
	default:
		return nil, ErrInvalidEndpoint
	}
}

type doOptions struct {
	data    interface{}
	context context.Context
	headers map[string]string
}

func (c *Client) do(method, path string, doOptions doOptions) (*http.Response, error) {
	var body io.Reader
	if doOptions.data != nil {
		buf, err := json.Marshal(doOptions.data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(buf)
	}

	protocol := c.endpointURL.Scheme
	var u string
	switch protocol {
	case unixProtocol:
	default:
		u = c.getURL(path)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	if doOptions.data != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if method == http.MethodPost {
		req.Header.Set("Content-Type", "plain/text")
	}

	for k, v := range doOptions.headers {
		req.Header.Set(k, v)
	}

	ctx := doOptions.context
	if ctx == nil {
		ctx = context.Background()
	}
	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, ErrConnectionRefused
		}
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, err
	}
	return resp, nil
}

func (c *Client) getURL(path string) string {
	return ""
}
