package dockerclient

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewTSLClient(t *testing.T) {
	type field struct {
		endpoint       string
		dockerCertPath string
	}
	type want struct {
		endpoint string
	}
	tests := []struct {
		name  string
		field field
		want  want
	}{
		{
			name: "tcp",
			field: field{
				endpoint:       "tcp://192.168.57.3:2376",
				dockerCertPath: "testing/data/",
			},
			want: want{
				endpoint: "tcp://192.168.57.3:2376",
			},
		},
		{
			name: "no protocol",
			field: field{
				endpoint:       "192.168.57.3:2376",
				dockerCertPath: "testing/data/",
			},
			want: want{
				endpoint: "tcp://192.168.57.3:2376",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//got, err := NewClient(tt.field.endpoint, tt.field.dockerCertPath)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//if dhost := got.HTTPClient.DaemonHost(); dhost != tt.want.endpoint {
			//	t.Errorf("NewClient(tt.field.endpoint, tt.field.dockerCertPath) = %v, want %v", got, tt.want)
			//}
		})
	}
}

type FakeRoundTripper struct {
	message  string
	status   int
	header   map[string]string
	requests []*http.Request
}

func (rt *FakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	body := strings.NewReader(rt.message)
	rt.requests = append(rt.requests, r)
	res := &http.Response{
		StatusCode: rt.status,
		Body:       io.NopCloser(body),
		Header:     make(http.Header),
	}
	for k, v := range rt.header {
		res.Header.Set(k, v)
	}
	return res, nil
}
