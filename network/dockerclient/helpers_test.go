package dockerclient

import (
	"errors"
	"testing"
)

func expectNoFoundContainer(t *testing.T, id string, err error) {
	t.Helper()
	var containerErr *ContainerNotRunning
	if !errors.As(err, &containerErr) {
		t.Fatalf("Container: Wrong error information. Want %v. Got %v.", containerErr, err)
	}
	if containerErr.ID != id {
		t.Errorf("Container: Wrong container in error Want %q. Got %q", id, containerErr.ID)
	}
}
