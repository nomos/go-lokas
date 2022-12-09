package dockerclient

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRemoveContainer(t *testing.T) {
	t.Parallel()
	fakeRT := &FakeRoundTripper{message: "", status: http.StatusOK}
	client := newTestClient(fakeRT)
	id := "guowei"
	opts := RemoveContainerOptions{ID: id}
	err := client.RemoveContainer(opts)
	if err != nil {
		t.Fatal(err)
	}
	req := fakeRT.requests[0]
	if req.Method != http.MethodDelete {
		t.Errorf("RemoveContainer: Wrong HTTP method. Want %q. Got %q.", http.MethodDelete, req.Method)
	}
	expectedURL, _ := url.Parse(client.getLocalURL("/containers/" + id))
	if gotPath := req.URL.Path; gotPath != expectedURL.Path {
		t.Errorf("RemoveContainer Wrong path in request. Want %q. Got %q.", expectedURL, gotPath)
	}
}

func TestRemoveContainerNotFound(t *testing.T) {
	t.Parallel()
	client := newTestClient(&FakeRoundTripper{message: "No such container", status: http.StatusNotFound})
	id := "guowei"
	err := client.RemoveContainer(RemoveContainerOptions{ID: id})
	expectNoFoundContainer(t, id, err)
}
