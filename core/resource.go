package core

import (
	"bytes"
	"encoding/json"
	"html/template"
)

//Namespace 命名空间
type Namespace struct {
	Name string `json:"name"`
}

//String ...
func (ns *Namespace) String() string {
	d, _ := json.Marshal(ns)
	return string(d)

}

func (ns *Namespace) Parse(s string) error {
	return json.Unmarshal([]byte(s), ns)
}

//Service 服务
type Service struct {
	Namespace string           `json:"namespace"`
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	Endpoints []Endpoint       `json:"endpoints"`
	Docker    ContainerService `json:"docker,omitempty"`
	RuntimeID string           `json:"-"`
}

//String ...
func (svc *Service) String() string {
	d, _ := json.Marshal(svc)
	return string(d)
}

func (svc *Service) Parse(s string) error {
	return json.Unmarshal([]byte(s), svc)
}

func (svc *Service) ParseTpl(hostname string, values map[string]interface{}) error {
	if svc.Kind == "docker" {
		tmpl := template.New("tpl")
		tmpl, err := tmpl.Parse(svc.Docker.String())
		if err != nil {
			return err
		}
		vars := ContainerValues{
			Global: values,
		}
		for _, e := range svc.Endpoints {
			if e.Hostname == hostname {
				vars.Endpoint = e
				break
			}
		}
		var b bytes.Buffer
		err = tmpl.Execute(&b, vars)
		if err != nil {
			return err
		}
		csvc := new(ContainerService)
		err = csvc.Parse(b.String())
		if err != nil {
			return err
		}
		svc.Docker = *csvc
	}
	return nil
}

//Endpoint 端点
type Endpoint struct {
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Status    string                 `json:"status,omitempty"`
	State     string                 `json:"state,omitempty"`
	Port      int                    `json:"port,omitempty"`
	Hostname  string                 `json:"hostname,omitempty"`
	HostIP    string                 `json:"hostIP,omitempty"`
	Created   int64                  `json:"created,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
	NodeInfo  interface{}            `json:"hostInfo,omitempty"`
	Updated   int64                  `json:"updated,omitempty"`
}

//String ...
func (e *Endpoint) String() string {
	d, _ := json.Marshal(e)
	return string(d)
}

func (e *Endpoint) Parse(s string) error {
	return json.Unmarshal([]byte(s), e)
}

//EndpointLogOpt 日志输出
type EndpointLogOpt struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Tail     string `json:"tail"`
	Since    string `json:"since"`
}

//Method 方法
type Method struct {
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
}

//String ...
func (m *Method) String() string {
	d, _ := json.Marshal(m)
	return string(d)
}

func (m *Method) Parse(s string) error {
	return json.Unmarshal([]byte(s), m)
}
