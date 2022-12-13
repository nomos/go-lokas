package dockerclient

import (
	"encoding/json"
	"errors"
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

type NoSuchService struct {
	ID  string
	Err error
}

func (err *NoSuchService) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return "No such service: " + err.ID
}

type UpdateServiceOptions struct {
	Version uint64
	swarm.ServiceSpec
}

func (c *Client) UpdateService(id string, opts UpdateServiceOptions) error {
	path := "/services/" + id + "/update?" + queryStringVersion(opts)
	resp, err := c.do(http.MethodPost, path, doOptions{
		data: opts.ServiceSpec,
	})
	defer resp.Body.Close()
	if err != nil {
		var e *Error
		if errors.As(err, &e) && e.Status == http.StatusNotFound {
			return &NoSuchService{ID: id}
		}
	}
	return nil
}
