package base

import (
	"github.com/gin-gonic/gin"
	"github.com/oars-sigs/oars-cloud/core"
)

//BaseController 基础controller
type BaseController struct {
	Mgr *core.APIManager
}

//Health 健康状态
func (c *BaseController) Health(ctx *gin.Context) {
	ctx.JSON(200, map[string]interface{}{"status": "ok"})
}
