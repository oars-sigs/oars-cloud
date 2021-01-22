package core

import (
	"context"
	"strings"
)

type APIManager struct {
	Cfg   *Config
	Admin ServiceInterface
}

type APIInput struct {
	Method string      `json:"method"`
	Args   interface{} `json:"args"`
}

func ParseServicePath(s string) (namespace, service, resource, action string) {
	items := strings.Split(s, ".")
	for i, item := range items {
		switch i {
		case 0:
			namespace = item
		case 1:
			service = item
		case 2:
			resource = item
		case 3:
			action = item
		}
	}
	return
}

type ServiceInterface interface {
	Call(ctx context.Context, resource, action string, args interface{}, reply *APIReply) error
}

//ServiceErrorCode 服务状态码
type ServiceErrorCode int

const (
	//ServiceSuccessCode 成功
	ServiceSuccessCode ServiceErrorCode = 10000 + iota
	//ServiceInternalErrorCode 内部错误
	ServiceInternalErrorCode
	//ServiceInvalidParameterCode 无效参数
	ServiceInvalidParameterCode
	//ServiceMethodNotFoundCode 方法不存在
	ServiceMethodNotFoundCode
	//ServiceResourceNotFoundCode 资源不存在
	ServiceResourceNotFoundCode
)

//GetErrorCodes 获取状态码表
func GetErrorCodes() map[string]ServiceErrorCode {
	return map[string]ServiceErrorCode{
		"ServiceSuccessCode":       ServiceSuccessCode,
		"ServiceInternalErrorCode": ServiceInternalErrorCode,
		"InvalidParameterCode":     ServiceInvalidParameterCode,
	}
}

//APIReply runtime reply
type APIReply struct {
	Code    ServiceErrorCode `json:"code"`
	Msg     string           `json:"msg,omitempty"`
	SubCode string           `json:"sub_code,omitempty"`
	SubMsg  string           `json:"sub_msg,omitempty"`
	Data    interface{}      `json:"data,omitempty"`
}

//NewAPIError runtime error
func NewAPIError(err error) *APIReply {
	return &APIReply{
		Code: ServiceSuccessCode,
		Msg:  err.Error(),
	}
}

//NewAPIReply runtime reply
func NewAPIReply(data interface{}) *APIReply {
	return &APIReply{
		Code: ServiceSuccessCode,
		Data: data,
	}
}
