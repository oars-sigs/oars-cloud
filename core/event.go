package core

import "encoding/json"

//Event 事件
type Event struct {
	*ResourceMeta
	Action  string `json:"action"`
	Status  string `json:"status"`
	From    string `json:"from"`
	Message string `json:"message"`
	Number  int64  `json:"number"`
}

const (
	//CreateEventAction 创建事件操作
	CreateEventAction = "create"
	//DeleteEventAction 删除事件操作
	DeleteEventAction = "delete"
	//StartEventAction 启动事件操作
	StartEventAction = "start"
	//ImagePullEventAction 拉镜像事件操作
	ImagePullEventAction = "imagePull"

	//SuccessEventStatus 成功事件
	SuccessEventStatus = "success"
	//FailEventStatus 失败事件
	FailEventStatus = "fail"
	//InProgressEventStatus 进行中事件
	InProgressEventStatus = "inProgress"
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
	return l.Name + "/" + l.Action + "/" + l.Status
}

//ResourcePrefixKey ...
func (l *Event) ResourcePrefixKey() string {
	if l.ResourceMeta == nil {
		return ""
	}
	if l.Status != "" {
		return l.Name + "/" + l.Action + "/" + l.Status
	}
	if l.Action != "" {
		return l.Name + "/" + l.Action + "/"
	}
	return l.Name + "/"
}

//GenName ...
func (l *Event) GenName(r Resource) {
	if l.ResourceMeta == nil {
		l.ResourceMeta = new(ResourceMeta)
	}
	l.ResourceMeta.Name = r.ResourceGroup() + "/" + r.ResourceKind() + "/" + r.ResourceKey()
}
