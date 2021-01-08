package core

import "encoding/json"

//Event 事件
type Event struct {
	*ResourceMeta
	From    string `json:"from"`
	Message string `json:"message"`
}

//String ...
func (l *Event) String() string {
	d, _ := json.Marshal(l)
	return string(d)
}

//Parse ...
func (l *Event) Parse(s string) error {
	return json.Unmarshal([]byte(s), l)
}

//New ...
func (l *Event) New() Resource {
	return &Event{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (l *Event) ResourceGroup() string {
	return "clusters"
}

//ResourceKind ...
func (l *Event) ResourceKind() string {
	return "event"
}

//ResourceKey ...
func (l *Event) ResourceKey() string {
	return "namespaces/" + l.Namespace + "/" + l.Name
}

//ResourcePrefixKey ...
func (l *Event) ResourcePrefixKey() string {
	if l.ResourceMeta == nil {
		return "namespaces/"
	}
	if l.Namespace != "" {
		return "namespaces/" + l.Namespace + "/" + l.Name
	}
	return "namespaces/"
}
