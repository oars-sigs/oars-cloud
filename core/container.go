package core

import "encoding/json"

//ContainerService 容器服务
type ContainerService struct {
	ID              string                 `json:"-"`
	Name            string                 `json:"-"`
	Labels          map[string]string      `json:"labels"`
	Image           string                 `json:"image,omitempty"`
	ImagePullPolicy string                 `json:"imagePullPolicy,omitempty"`
	ImagePullAuth   string                 `json:"imagePullAuth,omitempty"`
	Volumes         []string               `json:"volumes,omitempty"`
	DependsOn       []string               `json:"depends_on,omitempty"`
	Environment     []string               `json:"environment,omitempty"`
	Restart         string                 `json:"restart,omitempty"`
	Command         StrSlice               `json:"command,omitempty"`
	Shell           StrSlice               `json:"shell,omitempty"`
	User            string                 `json:"user,omitempty"`
	Port            *ContainerPort         `json:"port,omitempty"`
	Resources       *ContainerResource     `json:"resource,omitempty"`
	CapAdd          StrSlice               `json:"cap_add,omitempty"`
	CapDrop         StrSlice               `json:"cap_drop,omitempty"`
	Devices         []string               `json:"devices,omitempty"`
	Entrypoint      StrSlice               `json:"entrypoint,omitempty"`
	ExtraHosts      []string               `json:"extra_hosts,omitempty"`
	NetworkMode     string                 `json:"network_mode,omitempty"` //默认host network
	SecurityOpt     []string               `json:"security_opt,omitempty"`
	StopSignal      string                 `json:"stop_signal,omitempty"`
	Sysctls         map[string]string      `json:"sysctls,omitempty"`
	Ulimits         map[string]interface{} `json:"ulimits,omitempty"` //未实现
	Pid             string                 `json:"pid,omitempty"`
	Privileged      bool                   `json:"privileged,omitempty"`
	WorkingDir      string                 `json:"working_dir,omitempty"`
	ConfigMap       map[string]string      `json:"configmap,omitempty"`
	Ports           []string               `json:"ports,omitempty"`
	Expose          []string               `json:"expose,omitempty"`
	DNS             []string               `json:"dns,omitempty"`
}

var (
	ImagePullAlways       = "Always"
	ImagePullNever        = "Never"
	ImagePullIfNotPresent = "IfNotPresent"
)

//String ...
func (svc *ContainerService) String() string {
	d, _ := json.Marshal(svc)
	return string(d)
}

func (svc *ContainerService) Parse(s string) error {
	return json.Unmarshal([]byte(s), svc)
}

//ContainerPort 容器端口
type ContainerPort struct {
	Proxy         bool   `json:"proxy,omitempty"`
	ContainerPort int    `json:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type ContainerResource struct {
	CPU    float64 `json:"cpu,omitempty"`
	Memory int64   `json:"memory,omitempty"`
}

// StrSlice represents a string or an array of strings.
// We need to override the json decoder to accept both options.
type StrSlice []string

// UnmarshalJSON decodes the byte slice whether it's a string or an array of
// strings. This method is needed to implement json.Unmarshaler.
func (e *StrSlice) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	p := make([]string, 0, 1)
	if err := json.Unmarshal(b, &p); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		p = append(p, s)
	}

	*e = p
	return nil
}
