package core

import "encoding/json"

//Method 方法
type Method struct {
	*ResourceMeta
	ServiceName string `json:"serviceName"`
	Kind        string `json:"kind"`
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
