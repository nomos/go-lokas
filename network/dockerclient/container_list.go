package dockerclient

import (
	"encoding/json"
	"net/http"
)

// ListContainers 返回容器切片
//
// More: https://docs.docker.com/engine/api/v1.41/#tag/Container
func (c *Client) ListContainers() ([]APIContainers, error) {
	path := "/containers/json"
	resp, err := c.do(http.MethodGet, path, doOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var containers []APIContainers
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}
	return containers, nil
}
