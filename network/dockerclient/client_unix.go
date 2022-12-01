package dockerclient

import (
	"context"
	"net"
	"net/http"
)

const defaultHost = "unix:///var/run/docker.sock"

// initializeLocalClient initializes the local Unix domain socket client on Unix-style OS.
func (c *Client) initializeLocalClient(trFunc func() *http.Transport) {
	if c.endpointURL.Scheme != unixProtocol {
		return
	}
	sockPath := c.endpointURL.Path
	tr := trFunc()
	tr.Proxy = nil
	tr.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return c.Dialer.Dial(unixProtocol, sockPath)
	}
	c.HTTPClient.Transport = tr
}
