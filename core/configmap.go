package core

import "encoding/json"

//ConfigMap 配置
type ConfigMap struct {
	*ResourceMeta
	Data map[string]string `json:"data"`
}

//String ...
func (e *ConfigMap) String() string {
	d, _ := json.Marshal(e)
	return string(d)
}

//Parse ...
func (e *ConfigMap) Parse(s string) error {
	return json.Unmarshal([]byte(s), e)
}

//New ...
func (e *ConfigMap) New() Resource {
	return &ConfigMap{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (e *ConfigMap) ResourceGroup() string {
	return "configs"
}

//ResourceKind ...
func (e *ConfigMap) ResourceKind() string {
	return "configmap"
}

//ResourceKey ...
func (e *ConfigMap) ResourceKey() string {
	return "namespaces/" + e.Namespace + "/" + e.Name
}

//ResourcePrefixKey ...
func (e *ConfigMap) ResourcePrefixKey() string {
	if e.ResourceMeta == nil {
		return "namespaces/"
	}
	if e.Namespace != "" {
		return "namespaces/" + e.Namespace + "/" + e.Name
	}
	return "namespaces/"
}
