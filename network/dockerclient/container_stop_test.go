package dockerclient

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestStopContainer(t *testing.T) {
	t.Parallel()
	fakeRT := &FakeRoundTripper{message: "", status: http.StatusNoContent}
	client := newTestClient(fakeRT)
	id := "guowei"
	err := client.StopContainer(id)
	if err != nil {
		t.Fatal(err)
	}
	req := fakeRT.requests[0]
	if req.Method != http.MethodPost {
		t.Errorf("StopContainer: Wrong HTTP method. Want %q. Got %q.", http.MethodPost, req.Method)
	}
	expectedURL, _ := url.Parse(client.getLocalURL("/containers/" + id + "/stop"))
	if gotPath := req.URL.Path; gotPath != expectedURL.Path {
		t.Errorf("StopContainer: Wrong path in request. Want %q. Got %q.", expectedURL.Path, gotPath)
	}
}

func TestStopContainerNotFound(t *testing.T) {
	t.Parallel()
	client := newTestClient(&FakeRoundTripper{message: "no such container", status: http.StatusNotFound})
	id := "guowei"
	err := client.StopContainer(id)
	expectNoFoundContainer(t, id, err)
}

func TestStopContainerNotRunning(t *testing.T) {
	t.Parallel()
	client := newTestClient(&FakeRoundTripper{message: "container not running", status: http.StatusNotModified})
	id := "guowei"
	err := client.StopContainer(id)
	expected := &ContainerNotRunning{ID: id}
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("StopContainer: Wrong error returned. Want %v. Got %v.", expected, err)
	}
}
