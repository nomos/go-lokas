package dockerclient

import (
	"testing"
)

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
			got, err := NewTLSClient(tt.field.endpoint, tt.field.dockerCertPath)
			if err != nil {
				t.Fatal(err)
			}
			if dhost := got.DaemonHost(); dhost != tt.want.endpoint {
				t.Errorf("NewTLSClient(tt.field.endpoint, tt.field.dockerCertPath) = %v, want %v", got, tt.want)
			}
		})
	}
}
