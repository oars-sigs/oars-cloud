package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
)

const (
	DockerServiceKind  = "docker"
	StaticServiceKind  = "static"
	RuntimeServiceKind = "runtime"
)

//Service 服务
type Service struct {
	*ResourceMeta
	Kind          string            `json:"kind"`
	Endpoints     []ServiceEndpoint `json:"endpoints"`
	Docker        ContainerService  `json:"docker,omitempty"`
	Static        *StaticService    `json:"static,omitempty"`
	VirtualServer *VirtualServer    `json:"vs,omitempty"`
}

//ServiceEndpoint 服务端点
type ServiceEndpoint struct {
	Name     string                 `json:"name"`
	Hostname string                 `json:"hostname,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Domain   string                 `json:"domain,omitempty"`
	IP       string                 `json:"ip,omitempty"`
}

type StaticService struct {
	Endpoints []ServiceEndpoint `json:"endpoints"`
}

//VirtualServer Linux Virtual Server
type VirtualServer struct {
	ClusterIP string   `json:"clusterIP,omitempty"`
	Ports     []string `json:"ports"`
}

//ServiceValues 配置参数
type ServiceValues struct {
	Global   map[string]interface{}
	Endpoint ServiceEndpoint
	Node     Node
}

//String ...
func (svc *Service) String() string {
	d, _ := json.Marshal(svc)
	return string(d)
}

//Parse ...
func (svc *Service) Parse(s string) error {
	return json.Unmarshal([]byte(s), svc)
}

//New ...
func (svc *Service) New() Resource {
	return &Service{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (svc *Service) ResourceGroup() string {
	return "services"
}

//ResourceKind ...
func (svc *Service) ResourceKind() string {
	return "svc"
}

//ResourceKey ...
func (svc *Service) ResourceKey() string {
	return "namespaces/" + svc.Namespace + "/" + svc.Name
}

//ResourcePrefixKey ...
func (svc *Service) ResourcePrefixKey() string {
	if svc.ResourceMeta == nil {
		return "namespaces/"
	}
	if svc.Namespace != "" {
		return "namespaces/" + svc.Namespace + "/" + svc.Name
	}
	return "namespaces/"
}

//ParseContainer ...
func (svc *Service) ParseContainer(vars ServiceValues) (*ContainerService, error) {
	if svc.Kind == "docker" {
		tmpl := template.New("tpl")
		tmpl, err := tmpl.Parse(svc.Docker.String())
		if err != nil {
			return nil, err
		}
		var b bytes.Buffer
		err = tmpl.Execute(&b, vars)
		if err != nil {
			return nil, err
		}
		csvc := new(ContainerService)
		err = csvc.Parse(b.String())
		if err != nil {
			return nil, err
		}
		return csvc, nil
	}
	return nil, errors.New("not support kind")
}
