package dockerclient

import (
	"encoding/json"
	"github.com/docker/docker/api/types/swarm"
	"net/http"
)

func (c *Client) ListServices() ([]swarm.Service, error) {
	var services []swarm.Service
	path := "/services?"
	resp, err := c.do(http.MethodGet, path, doOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}
	return services, nil
}
