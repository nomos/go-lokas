package dockerclient

import (
	"encoding/json"
	"github.com/docker/docker/api/types/swarm"
	"net/http"
)

// ListServices returns a slices of services
//
// More: https://docs.docker.com/engine/api/v1.41/#tag/Service/operation/ServiceList
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
