package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/server/apis/base"
	v1 "github.com/oars-sigs/oars-cloud/pkg/server/apis/v1"
)

func NewV1(r *gin.Engine, mgr *core.APIManager) {
	basec := &base.BaseController{Mgr: mgr}
	gatewayc := &v1.GatewayController{BaseController: basec}
	r.GET("/health", basec.Health)
	apiv1 := r.Group("/api")
	apiv1.POST("gateway", gatewayc.Gateway)
	apiv1.GET("exec/:hostname/:id", gatewayc.Exec)
}
