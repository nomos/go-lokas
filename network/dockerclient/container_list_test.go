package dockerclient

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestListContainers(t *testing.T) {
	t.Parallel()
	jsonContainers := `[
  {
    "Id": "8dfafdbc3a40",
    "Names": [
      "/boring_feynman"
    ],
    "Image": "ubuntu:latest",
    "ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82",
    "Command": "echo 1",
    "Created": 1367854155,
    "State": "Exited",
    "Status": "Exit 0",
    "Ports": [
      {
        "PrivatePort": 2222,
        "PublicPort": 3333,
        "Type": "tcp"
      }
    ],
    "Labels": {
      "com.example.vendor": "Acme",
      "com.example.license": "GPL",
      "com.example.version": "1.0"
    },
    "SizeRw": 12288,
    "SizeRootFs": 0,
    "HostConfig": {
      "NetworkMode": "default"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "NetworkID": "7ea29fc1412292a2d7bba362f9253545fecdfa8ce9a6e37dd10ba8bee7129812",
          "EndpointID": "2cdc4edb1ded3631c81f57966563e5c8525b81121bb3706a9a9a3ae102711f3f",
          "Gateway": "172.17.0.1",
          "IPAddress": "172.17.0.2",
          "IPPrefixLen": 16,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "MacAddress": "02:42:ac:11:00:02"
        }
      }
    },
    "Mounts": [
      {
        "Name": "fac362...80535",
        "Source": "/data",
        "Destination": "/data",
        "Driver": "local",
        "Mode": "ro,Z",
        "RW": false,
        "Propagation": ""
      }
    ]
  },
  {
    "Id": "9cd87474be90",
    "Names": [
      "/coolName"
    ],
    "Image": "ubuntu:latest",
    "ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82",
    "Command": "echo 222222",
    "Created": 1367854155,
    "State": "Exited",
    "Status": "Exit 0",
    "Ports": [],
    "Labels": {},
    "SizeRw": 12288,
    "SizeRootFs": 0,
    "HostConfig": {
      "NetworkMode": "default"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "NetworkID": "7ea29fc1412292a2d7bba362f9253545fecdfa8ce9a6e37dd10ba8bee7129812",
          "EndpointID": "88eaed7b37b38c2a3f0c4bc796494fdf51b270c2d22656412a2ca5d559a64d7a",
          "Gateway": "172.17.0.1",
          "IPAddress": "172.17.0.8",
          "IPPrefixLen": 16,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "MacAddress": "02:42:ac:11:00:08"
        }
      }
    },
    "Mounts": []
  },
  {
    "Id": "3176a2479c92",
    "Names": [
      "/sleepy_dog"
    ],
    "Image": "ubuntu:latest",
    "ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82",
    "Command": "echo 3333333333333333",
    "Created": 1367854154,
    "State": "Exited",
    "Status": "Exit 0",
    "Ports": [],
    "Labels": {},
    "SizeRw": 12288,
    "SizeRootFs": 0,
    "HostConfig": {
      "NetworkMode": "default"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "NetworkID": "7ea29fc1412292a2d7bba362f9253545fecdfa8ce9a6e37dd10ba8bee7129812",
          "EndpointID": "8b27c041c30326d59cd6e6f510d4f8d1d570a228466f956edf7815508f78e30d",
          "Gateway": "172.17.0.1",
          "IPAddress": "172.17.0.6",
          "IPPrefixLen": 16,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "MacAddress": "02:42:ac:11:00:06"
        }
      }
    },
    "Mounts": []
  },
  {
    "Id": "4cb07b47f9fb",
    "Names": [
      "/running_cat"
    ],
    "Image": "ubuntu:latest",
    "ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82",
    "Command": "echo 444444444444444444444444444444444",
    "Created": 1367854152,
    "State": "Exited",
    "Status": "Exit 0",
    "Ports": [],
    "Labels": {},
    "SizeRw": 12288,
    "SizeRootFs": 0,
    "HostConfig": {
      "NetworkMode": "default"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "NetworkID": "7ea29fc1412292a2d7bba362f9253545fecdfa8ce9a6e37dd10ba8bee7129812",
          "EndpointID": "d91c7b2f0644403d7ef3095985ea0e2370325cd2332ff3a3225c4247328e66e9",
          "Gateway": "172.17.0.1",
          "IPAddress": "172.17.0.5",
          "IPPrefixLen": 16,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "MacAddress": "02:42:ac:11:00:05"
        }
      }
    },
    "Mounts": []
  }
]`
	var excepted []APIContainers
	err := json.Unmarshal([]byte(jsonContainers), &excepted)
	if err != nil {
		t.Fatal(err)
	}
	client := newTestClient(&FakeRoundTripper{message: jsonContainers, status: http.StatusOK})
	containers, err := client.ListContainers()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(excepted, containers) {
		t.Errorf("ListContainers: Expected %v. Got %v", excepted, containers)
	}
}
