package core

import "encoding/json"

//Cron 定时任务
type Cron struct {
	*ResourceMeta
	Expr     string      `json:"expr"`
	Disabled bool        `json:"disabled"`
	Service  *Service    `json:"service"`
	Status   *CronStatus `json:"status"`
}

type CronStatus struct {
	Next int64 `json:"next"`
	Prev int64 `json:"prev"`
}

//String ...
func (e *Cron) String() string {
	d, _ := json.Marshal(e)
	return string(d)
}

//Parse ...
func (e *Cron) Parse(s string) error {
	return json.Unmarshal([]byte(s), e)
}

//New ...
func (e *Cron) New() Resource {
	return &Cron{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (e *Cron) ResourceGroup() string {
	return "cluster"
}

//ResourceKind ...
func (e *Cron) ResourceKind() string {
	return "cron"
}

//ResourceKey ...
func (e *Cron) ResourceKey() string {
	return "namespaces/" + e.Namespace + "/" + e.Name
}

//ResourcePrefixKey ...
func (e *Cron) ResourcePrefixKey() string {
	if e.ResourceMeta == nil {
		return "namespaces/"
	}
	if e.Namespace != "" {
		return "namespaces/" + e.Namespace + "/"
	}
	return "namespaces/"
}
