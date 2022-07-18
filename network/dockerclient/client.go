package dockerclient

import (
	"errors"
	"github.com/docker/docker/client"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// ErrInvalidEndpoint is returned when the endpoint is not a valid URL.
	ErrInvalidEndpoint = errors.New("invalid endpoint")
)

type TLSClient struct {
	*client.Client
}

// NewTLSClient returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates, using a specific remote API version.
func NewTLSClient(endpoint string, dockerCertPath string) (*TLSClient, error) {
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

	endpoint, err := parseEndPoint(endpoint)
	if err != nil {
		return nil, err
	}

	if err := os.Setenv("DOCKER_CERT_PATH", dockerCertPath); err != nil {
		return nil, err
	}
	if err := os.Setenv("DOCKER_HOST", endpoint); err != nil {
		return nil, err
	}
	//TODO Docker version

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	ret := &TLSClient{cli}
	return ret, nil
}

func parseEndPoint(endpoint string) (string, error) {
	if endpoint != "" && !strings.Contains(endpoint, "://") {
		endpoint = "tcp://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint, ErrInvalidEndpoint
	}

	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		var e *net.AddrError
		if errors.As(err, &e) {
			if e.Err == "missing port in address" {
				return endpoint, ErrInvalidEndpoint
			}
		}

		number, err := strconv.ParseInt(port, 10, 64)
		if err != nil || number < 0 || number > 65536 {
			return endpoint, ErrInvalidEndpoint
		}
	}
	return endpoint, nil
}
