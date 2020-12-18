package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
)

//Service 服务
type Service struct {
	Version         string            `json:"version"`
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Endpoints       []ServiceEndpoint `json:"endpoints"`
	CurrentEndpoint *ServiceEndpoint  `json:"-"` //use to runtime
	Docker          ContainerService  `json:"docker,omitempty"`
	RuntimeID       string            `json:"-"`
	Created         int64             `json:"created,omitempty"`
	Updated         int64             `json:"updated,omitempty"`
}

//ServiceEndpoint 服务端点
type ServiceEndpoint struct {
	Name     string                 `json:"name"`
	Hostname string                 `json:"hostname"`
	Config   map[string]interface{} `json:"config"`
	Domain   string                 `json:"domain,omitempty"`
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
	return new(Service)
}

//ID ...
func (svc *Service) ID() string {
	return svc.Namespace + "/" + svc.Name
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

//ServiceStore endpiont store
type ServiceStore interface {
	List(ctx context.Context, arg *Service, opts *ListOptions) ([]*Service, error)
	Get(ctx context.Context, arg *Service, opts *GetOptions) (*Service, error)
	Put(ctx context.Context, arg *Service, opts *PutOptions) (*Service, error)
	Delete(ctx context.Context, arg *Service, opts *DeleteOptions) error
}
