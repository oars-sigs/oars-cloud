package core

import "encoding/json"

//Certificate 证书
type Certificate struct {
	*ResourceMeta
	Host  string `json:"host"`
	CA    string `json:"ca,omitempty"`
	CAKey string `json:"caKey,omitempty"`
	Cert  string `json:"cert,omitempty"`
	Key   string `json:"key,omitempty"`
}

//String ...
func (r *Certificate) String() string {
	d, _ := json.Marshal(r)
	return string(d)
}

//Parse ...
func (r *Certificate) Parse(s string) error {
	return json.Unmarshal([]byte(s), r)
}

//New ...
func (r *Certificate) New() Resource {
	return &Certificate{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (r *Certificate) ResourceGroup() string {
	return "ingresses"
}

//ResourceKind ...
func (r *Certificate) ResourceKind() string {
	return "cert"
}

//ResourceKey ...
func (r *Certificate) ResourceKey() string {
	return r.Name
}

//ResourcePrefixKey ...
func (r *Certificate) ResourcePrefixKey() string {
	if r.ResourceMeta == nil {
		return ""
	}
	return r.Name
}
