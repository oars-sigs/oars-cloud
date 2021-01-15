package v1

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oars-sigs/oars-cloud/pkg/server/apis/base"
)

//ProxyController ...
type ProxyController struct {
	*base.BaseController
}

func (c *ProxyController) Proxy(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	service := ctx.Param("service")
	protocol := ctx.Param("protocol")
	port := ctx.Param("port")
	proxy(namespace, service, protocol, port).ServeHTTP(ctx.Writer, ctx.Request)
}

func proxy(namespace, service, protocol, port string) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			addr := protocol + "://" + service + "." + namespace
			if port != "443" && port != "" && protocol == "https" {
				addr += ":" + port
			}
			if port != "80" && port != "" && protocol == "http" {
				addr += ":" + port
			}
			prefix := "/proxy/" + namespace + "/" + service + "/" + protocol + "/" + port
			destPath := strings.TrimPrefix(req.URL.Path, prefix)
			destURL, _ := url.Parse(addr + destPath)
			destURL.RawQuery = req.URL.RawQuery
			req.URL = destURL
			req.Host = destURL.Host
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
