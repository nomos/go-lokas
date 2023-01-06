package dockerclient

import (
	"encoding/json"
	"errors"
	"github.com/docker/docker/api/types/swarm"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestListServices(t *testing.T) {
	t.Parallel()
	jsonServices := `[
  {
    "ID": "9mnpnzenvg8p8tdbtq4wvbkcz",
    "Version": {
      "Index": 19
    },
    "CreatedAt": "2016-06-07T21:05:51.880065305Z",
    "UpdatedAt": "2016-06-07T21:07:29.962229872Z",
    "Spec": {
      "Name": "hopeful_cori",
      "TaskTemplate": {
        "ContainerSpec": {
          "Image": "redis"
        },
        "Resources": {
          "Limits": {},
          "Reservations": {}
        },
        "RestartPolicy": {
          "Condition": "any",
          "MaxAttempts": 0
        },
        "Placement": {},
        "ForceUpdate": 0
      },
      "Mode": {
        "Replicated": {
          "Replicas": 1
        }
      },
      "UpdateConfig": {
        "Parallelism": 1,
        "Delay": 1000000000,
        "FailureAction": "pause",
        "Monitor": 15000000000,
        "MaxFailureRatio": 0.15
      },
      "RollbackConfig": {
        "Parallelism": 1,
        "Delay": 1000000000,
        "FailureAction": "pause",
        "Monitor": 15000000000,
        "MaxFailureRatio": 0.15
      },
      "EndpointSpec": {
        "Mode": "vip",
        "Ports": [
          {
            "Protocol": "tcp",
            "TargetPort": 6379,
            "PublishedPort": 30001
          }
        ]
      }
    },
    "Endpoint": {
      "Spec": {
        "Mode": "vip",
        "Ports": [
          {
            "Protocol": "tcp",
            "TargetPort": 6379,
            "PublishedPort": 30001
          }
        ]
      },
      "Ports": [
        {
          "Protocol": "tcp",
          "TargetPort": 6379,
          "PublishedPort": 30001
        }
      ],
      "VirtualIPs": [
        {
          "NetworkID": "4qvuz4ko70xaltuqbt8956gd1",
          "Addr": "10.255.0.2/16"
        },
        {
          "NetworkID": "4qvuz4ko70xaltuqbt8956gd1",
          "Addr": "10.255.0.3/16"
        }
      ]
    }
  }
]`
	var excepted []swarm.Service
	err := json.Unmarshal([]byte(jsonServices), &excepted)
	if err != nil {
		t.Fatal(err)
	}
	client := newTestClient(&FakeRoundTripper{message: jsonServices, status: http.StatusOK})
	services, err := client.ListServices()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(services, excepted) {
		t.Errorf("ListService: Excepted %v. Got %v", excepted, services)
	}
}

func TestUpdateService(t *testing.T) {
	t.Parallel()
	fakeRT := &FakeRoundTripper{message: "", status: http.StatusOK}
	client := newTestClient(fakeRT)
	id := "guowei"
	update := UpdateServiceOptions{Version: 1}
	err := client.UpdateService(id, update)
	if err != nil {
		t.Fatal(err)
	}
	req := fakeRT.requests[0]
	if req.Method != http.MethodPost {
		t.Errorf("UpdateService: Wrong HTTP method. Want %q. Got %q.", http.MethodPost, req.Method)
	}
	expectedURL, _ := url.Parse(client.getLocalURL("/services/" + id + "/update?version=1"))
	if gotURI := req.URL.RequestURI(); gotURI != expectedURL.RequestURI() {
		t.Errorf("UpdateService: Wrong path in request. Want %q. Got %q.", expectedURL.RequestURI(), gotURI)
	}

	expectedContentType := "application/json"
	if contentType := req.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("UpdateService: Wrong content-type in request. Want %q. Got %q.", expectedContentType, contentType)
	}

	var out UpdateServiceOptions
	if err := json.NewDecoder(req.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	update.Version = 0
	if !reflect.DeepEqual(out, update) {
		t.Errorf("UpdateService: wrong body got %v want %v", out, update)
	}
}

func TestInspectService(t *testing.T) {
	t.Parallel()
	jsonService := `
{
  "ID": "9mnpnzenvg8p8tdbtq4wvbkcz",
  "Version": {
    "Index": 19
  },
  "CreatedAt": "2016-06-07T21:05:51.880065305Z",
  "UpdatedAt": "2016-06-07T21:07:29.962229872Z",
  "Spec": {
    "Name": "hopeful_cori",
    "TaskTemplate": {
      "ContainerSpec": {
        "Image": "redis"
      },
      "Resources": {
        "Limits": {},
        "Reservations": {}
      },
      "RestartPolicy": {
        "Condition": "any",
        "MaxAttempts": 0
      },
      "Placement": {},
      "ForceUpdate": 0
    },
    "Mode": {
      "Replicated": {
        "Replicas": 1
      }
    },
    "UpdateConfig": {
      "Parallelism": 1,
      "Delay": 1000000000,
      "FailureAction": "pause",
      "Monitor": 15000000000,
      "MaxFailureRatio": 0.15
    },
    "RollbackConfig": {
      "Parallelism": 1,
      "Delay": 1000000000,
      "FailureAction": "pause",
      "Monitor": 15000000000,
      "MaxFailureRatio": 0.15
    },
    "EndpointSpec": {
      "Mode": "vip",
      "Ports": [
        {
          "Protocol": "tcp",
          "TargetPort": 6379,
          "PublishedPort": 30001
        }
      ]
    }
  },
  "Endpoint": {
    "Spec": {
      "Mode": "vip",
      "Ports": [
        {
          "Protocol": "tcp",
          "TargetPort": 6379,
          "PublishedPort": 30001
        }
      ]
    },
    "Ports": [
      {
        "Protocol": "tcp",
        "TargetPort": 6379,
        "PublishedPort": 30001
      }
    ],
    "VirtualIPs": [
      {
        "NetworkID": "4qvuz4ko70xaltuqbt8956gd1",
        "Addr": "10.255.0.2/16"
      },
      {
        "NetworkID": "4qvuz4ko70xaltuqbt8956gd1",
        "Addr": "10.255.0.3/16"
      }
    ]
  }
}
`
	var expected swarm.Service
	if err := json.Unmarshal([]byte(jsonService), &expected); err != nil {
		t.Fatal(err)
	}
	fakeRT := &FakeRoundTripper{message: jsonService, status: http.StatusOK}
	client := newTestClient(fakeRT)
	id := "guowei"
	service, err := client.InspectService(id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*service, expected) {
		t.Errorf("InspectService: Excpected %v. Got %v.", expected, service)
	}
	expectedURL, err := url.Parse(client.getLocalURL("/services/guowei"))
	if gotPath := fakeRT.requests[0].URL.Path; gotPath != expectedURL.Path {
		t.Errorf("InspectService: Wrong path in request. Want %q. Got %q.", expectedURL.Path, gotPath)
	}
}

func TestInspectServiceNotFound(t *testing.T) {
	t.Parallel()
	client := newTestClient(&FakeRoundTripper{message: "no such service", status: http.StatusNotFound})
	service, err := client.InspectService("notfound")
	if service != nil {
		t.Errorf("InspectService: Expected <nil> service, got %#v", service)
	}
	expected := &NoSuchService{ID: "notfound"}
	var e *NoSuchService
	if errors.As(err, &e) && e.ID != expected.ID {
		t.Errorf("InspectService: Wrong error returned. Want %v. Got %v.", expected, err)
	}
}
