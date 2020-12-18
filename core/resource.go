package core

import (
	"encoding/json"
)

type Resource interface {
	String() string
	Parse(s string) error
	New() Resource
	ID() string
}

//Method 方法
type Method struct {
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Created     int64  `json:"created,omitempty"`
	Updated     int64  `json:"updated,omitempty"`
}

//String ...
func (m *Method) String() string {
	d, _ := json.Marshal(m)
	return string(d)
}

//Parse ...
func (m *Method) Parse(s string) error {
	return json.Unmarshal([]byte(s), m)
}
