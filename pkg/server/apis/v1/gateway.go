package v1

import (
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/server/apis/base"

	"github.com/gin-gonic/gin"
)

type GatewayController struct {
	*base.BaseController
}

func (c *GatewayController) Gateway(ctx *gin.Context) {
	var in core.APIInput
	err := ctx.ShouldBindJSON(&in)
	if err != nil {
		ctx.JSON(200, e.InvalidParameterError(err))
		return
	}

	ns, service, resource, action := core.ParseServicePath(in.Method)
	if ns == "" || service == "" || resource == "" {
		ctx.JSON(200, e.InvalidParameterError())
		return
	}
	if ns == "system" && service == "admin" {
		var reply core.APIReply
		err := c.Mgr.Admin.Call(ctx, resource, action, in.Args, &reply)
		if err != nil {
			ctx.JSON(200, e.InternalError(err))
			return
		}
		ctx.JSON(200, reply)
		return
	}
	ctx.JSON(200, e.MethodNotFoundError())
}
