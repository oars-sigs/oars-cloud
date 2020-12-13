package endpoints

import (
	"fmt"

	"github.com/oars-sigs/oars-cloud/core"
)

func getPrefixKey(selector *core.Endpoint) string {
	prefix := "services/endpoint/namespaces/"
	if selector.Namespace != "" {
		prefix += selector.Namespace + "/"
		if selector.Service != "" {
			prefix += selector.Service + "/"
		}
	}
	return prefix
}

func getKey(arg *core.Endpoint) string {
	return fmt.Sprintf("services/endpoint/namespaces/%s/%s/%s", arg.Namespace, arg.Service, arg.Name)
}

func isContain(src, selector *core.Endpoint) bool {
	if selector.Namespace == "" {
		return true
	}
	if selector.Namespace != "" && selector.Namespace == src.Namespace {
		if selector.Service == "" {
			return true
		}
		if selector.Service != "" && selector.Service == src.Service {
			if selector.Name == "" {
				return true
			}
			if selector.Name != "" && selector.Name == src.Name {
				return true
			}
		}
	}
	return false
}
