package dockerclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/homedir"
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

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type Client struct {
	Dialer      Dialer
	TLSConfig   *tls.Config
	HTTPClient  *http.Client
	endpointURL *url.URL
}

// NewClient returns a Client
func NewClient() (*Client, error) {
	dockerEnv, err := getDockerEnv()
	if err != nil {
		return nil, err
	}
	endpoint := dockerEnv.dockerHost
	if dockerEnv.dockerTLSVerify {
		//TODO: TLS Client
	}

	u, err := parseEndPoint(endpoint)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Dialer:      &net.Dialer{},
		endpointURL: u,
		HTTPClient:  defaultClient(),
	}
	c.initializeLocalClient(defaultTransport)
	return c, nil
}

type dockerEnv struct {
	dockerHost      string
	dockerCertPath  string
	dockerTLSVerify bool
}

func getDockerEnv() (*dockerEnv, error) {
	var err error
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = defaultHost
	}
	dockerTLSVerify := os.Getenv("DOCKER_TLS_VERIFY") != ""
	var dockerCertPath string
	if dockerTLSVerify {
		dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
		if dockerCertPath == "" {
			home := homedir.Get()
			if home == "" {
				return nil, errors.New("environment variable HOME must be set if DOCKER_CERT_PATH is not set")
			}
			dockerCertPath = filepath.Join(home, ".docker")
			dockerCertPath, err = filepath.Abs(dockerCertPath)
			if err != nil {
				return nil, err
			}
		}
	}
	return &dockerEnv{
		dockerHost:      dockerHost,
		dockerCertPath:  dockerCertPath,
		dockerTLSVerify: dockerTLSVerify,
	}, nil
}

// defaultClient returns a new http.Client
func defaultClient() *http.Client {
	return &http.Client{}
}

func defaultTransport() *http.Transport {
	return &http.Transport{}
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
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}
	return u, nil
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
	var u string
	u = c.getLocalURL(path)
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

// getLocalURL returns the URL needed to make an HTTP request over a UNIX.
func (c *Client) getLocalURL(path string) string {
	u := *c.endpointURL
	u.Scheme = "http"
	u.Host = "unix.sock"
	u.Path = ""
	urlStr := strings.TrimRight(u.String(), "/")
	return fmt.Sprintf("%s%s", urlStr, path)
}

// apiClientError represents failures in the API.
type apiClientError struct {
	Status  int
	Message string
}

func (e *apiClientError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.Status, e.Message)
}
