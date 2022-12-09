package dockerclient

import (
	"errors"
	"net/http"
)

// RemoveContainerOptions 封装用于删除容器的选项
type RemoveContainerOptions struct {
	// 容器ID
	ID string
}

// RemoveContainer 删除容器
//
// More: https://docs.docker.com/engine/api/v1.25/#tag/Container/operation/ContainerDelete
func (c *Client) RemoveContainer(opts RemoveContainerOptions) error {
	path := "/containers/" + opts.ID
	resp, err := c.do(http.MethodDelete, path, doOptions{})
	defer resp.Body.Close()
	if err != nil {
		var e *Error
		if errors.As(err, &e) && e.Status == http.StatusNotFound {
			return &ContainerNotFound{ID: opts.ID}
		}
		return err
	}
	return nil
}
