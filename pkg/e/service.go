package e

import (
	"github.com/oars-sigs/oars-cloud/core"
)

func getErr(err ...error) string {
	submsg := ""
	if len(err) > 0 {
		if err[0] == nil {
			return "unknow error"
		}
		submsg = err[0].Error()
	}
	return submsg
}

//InvalidParameterError 无效参数
func InvalidParameterError(err ...error) *core.APIReply {
	return &core.APIReply{
		Code:    core.ServiceInvalidParameterCode,
		Msg:     "参数错误",
		SubCode: "invalid-param",
		SubMsg:  getErr(err...),
	}
}

//InternalError 内部错误
func InternalError(err ...error) *core.APIReply {
	return &core.APIReply{
		Code:    core.ServiceInternalErrorCode,
		Msg:     "内部错误",
		SubCode: "unknow-error",
		SubMsg:  getErr(err...),
	}
}

//MethodNotFoundError 内部错误
func MethodNotFoundError(err ...error) *core.APIReply {
	return &core.APIReply{
		Code:    core.ServiceMethodNotFoundCode,
		Msg:     "方法不存在",
		SubCode: "method-not-found",
		SubMsg:  getErr(err...),
	}
}

//MethodNotFoundMethod ...
func MethodNotFoundMethod() *core.APIReply {
	return MethodNotFoundError()
}
