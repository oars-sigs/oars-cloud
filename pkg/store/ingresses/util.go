package ingresses

import (
	"fmt"

	"github.com/oars-sigs/oars-cloud/core"
)

func getListenerKey(selector *core.IngressListener) string {
	prefix := "ingresses/listener/"
	return prefix + selector.Name
}

func getRouteKey(selector *core.IngressRoute) string {
	prefix := fmt.Sprintf("ingresses/route/listener/%s/namespaces/%s/%s", selector.Namespace, selector.Listener, selector.Name)
	return prefix
}

func getRoutePrefixKey(selector *core.IngressRoute) string {
	prefix := "ingresses/route/listener/"
	if selector.Namespace != "" {
		prefix += selector.Namespace + "/"
		if selector.Listener != "" {
			prefix += selector.Listener + "/" + selector.Name
		}
	}
	return prefix
}
