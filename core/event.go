package core

import "encoding/json"

//Event 事件
type Event struct {
	*ResourceMeta
	Kind    string `json:"kind"`
	From    string `json:"from"`
	Message string `json:"message"`
}

const (
	//CreateEventKind 创建事件
	CreateEventKind = "create"
	//DeleteEventKind 删除事件
	DeleteEventKind = "delete"
)

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
	return l.Name + "/" + l.Kind
}

//ResourcePrefixKey ...
func (l *Event) ResourcePrefixKey() string {
	if l.ResourceMeta == nil {
		return ""
	}
	return l.Name
}
