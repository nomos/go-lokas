package dockerclient

import (
	"encoding/json"
	"net/http"
)

// APIImages represent an image returned to the ListImages call.
type APIImages struct {
	ID          string            `json:"Id"`
	RepoTags    []string          `json:"RepoTags,omitempty"`
	Created     int64             `json:"Created,omitempty"`
	Size        int64             `json:"Size,omitempty"`
	VirtualSize int64             `json:"VirtualSize,omitempty"`
	ParentID    string            `json:"ParentId,omitempty"`
	RepoDigests []string          `json:"RepoDigests,omitempty"`
	Labels      map[string]string `json:"Labels,omitempty"`
}

// ListImages returns the list of available images in the server.
//
// More: https://docs.docker.com/engine/api/v1.23/#list-images
func (c *Client) ListImages() ([]APIImages, error) {
	path := "/images/json"
	resp, err := c.do(http.MethodGet, path, doOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var images []APIImages
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, err
	}
	return images, nil
}
