package dockerclient

import (
	"errors"
	"fmt"
	"net/http"
)

// StopContainer 停止容器
func (c *Client) StopContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/stop", id)
	resp, err := c.do(http.MethodPost, path, doOptions{})
	if err != nil {
		var e *Error
		if errors.As(err, &e) && e.Status == http.StatusNotFound {
			return &ContainerNotFound{ID: id}
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return &ContainerNotRunning{ID: id}
	}
	return nil
}
