package dockerclient

// APIPort Docker API 返回的端口映射类型
type APIPort struct {
	Type        string `json:"Type,omitempty"`
	PublicPort  int64  `json:"publicPort,omitempty"`
	PrivatePort int64  `json:"privatePort,omitempty"`
}

// APIMount 容器的挂载点
type APIMount struct {
	Name        string `json:"Name,omitempty"`
	Mode        string `json:"Mode,omitempty"`
	Source      string `json:"Source,omitempty"`
	Driver      string `json:"Driver,omitempty"`
	Destination string `json:"Destination,omitempty"`
	Propagation string `json:"Propagation,omitempty"`
}

// ContainerNetwork 容器的网络
type ContainerNetwork struct {
	Aliases             []string `json:"Aliases,omitempty"`
	IPPrefixLen         int      `json:"IPPrefixLen,omitempty"`
	GlobalIPv6PrefixLen int      `json:"GlobalIPv6PrefixLen,omitempty"`
	Gateway             string   `json:"Gateway,omitempty"`
	NetworkID           string   `json:"NetworkID,omitempty"`
	IPAddress           string   `json:"IPAddress,omitempty"`
	EndpointID          string   `json:"EndpointID,omitempty"`
	MacAddress          string   `json:"MacAddress,omitempty"`
	IPv6Gateway         string   `json:"IPv6Gateway,omitempty"`
	GlobalIPv6Address   string   `json:"GlobalIPv6Address,omitempty"`
}

// NetworkList 容器的网络映射
type NetworkList struct {
	Networks map[string]ContainerNetwork `json:"Networks,omitempty"`
}

// APIContainers Docker API ListContainers 响应体切片中的容器实体
type APIContainers struct {
	ID         string            `json:"Id"`
	Image      string            `json:"Image,omitempty"`
	State      string            `json:"State,omitempty"`
	Status     string            `json:"Status,omitempty"`
	Command    string            `json:"Command,omitempty"`
	SizeRw     int64             `json:"SizeRw,omitempty"`
	SizeRootFs int64             `json:"SizeRootFs,omitempty"`
	Names      []string          `json:"Names,omitempty"`
	Ports      []APIPort         `json:"Ports,omitempty"`
	Mounts     []APIMount        `json:"Mounts,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty"`
	Networks   NetworkList       `json:"NetworkSettings,omitempty"`
}
