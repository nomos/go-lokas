package dockerclient

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func newTestClient(rt http.RoundTripper) Client {
	u, _ := parseEndPoint(defaultHost)
	client := Client{
		Dialer:      &net.Dialer{},
		HTTPClient:  &http.Client{Transport: rt},
		endpointURL: u,
	}
	return client
}

func TestListImages(t *testing.T) {
	t.Parallel()
	body := `[
	  {
		 "RepoTags": [
		   "ubuntu:12.04",
		   "ubuntu:precise",
		   "ubuntu:latest"
		 ],
		 "Id": "8dbd9e392a964056420e5d58ca5cc376ef18e2de93b5cc90e868a1bbc8318c1c",
		 "Created": 1365714795,
		 "Size": 131506275,
		 "VirtualSize": 131506275,
		 "Labels": {}
	  },
	  {
		 "RepoTags": [
		   "ubuntu:12.10",
		   "ubuntu:quantal"
		 ],
		 "ParentId": "27cf784147099545",
		 "Id": "b750fe79269d2ec9a3c593ef05b4332b1d1a02a62b4accb2c21d589ff2f5f2dc",
		 "Created": 1364102658,
		 "Size": 24653,
		 "VirtualSize": 180116135,
		 "Labels": {
			"com.example.version": "v1"
		 }
	  }
]`
	var expected []APIImages
	if err := json.Unmarshal([]byte(body), &expected); err != nil {
		t.Fatal(err)
	}

	client := newTestClient(&FakeRoundTripper{message: body, status: http.StatusOK})
	images, err := client.ListImages()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("ListImages: Wrong return value. Want %#v. Got %#v", expected, images)
	}
}

func TestRemoveImage(t *testing.T) {
	t.Parallel()
	name := "test"
	fakeRT := &FakeRoundTripper{message: "", status: http.StatusNoContent}
	client := newTestClient(fakeRT)
	err := client.RemoveImage(name)
	if err != nil {
		t.Fatal(err)
	}
	req := fakeRT.requests[0]
	expectedMethod := http.MethodDelete
	if req.Method != expectedMethod {
		t.Errorf("RemoveImage: Wrong HTTP method. Want %s. Got %s.", expectedMethod, req.Method)
	}
	u, _ := url.Parse(client.getLocalURL("/images/" + name))
	if req.URL.Path != u.Path {
		t.Errorf("RemoveImage: Wrong request path. Want %q. Got %q.", u.Path, req.URL.Path)
	}
}

func TestRemoveImageNotFound(t *testing.T) {
	t.Parallel()
	client := newTestClient(&FakeRoundTripper{message: "no such image", status: http.StatusNotFound})
	err := client.RemoveImage("guowei")
	if !errors.Is(err, ErrNoSuchImage) {
		t.Errorf("RemoveImage: Wrong err. Want %q. Got %q.", ErrNoSuchImage, err)
	}
}
