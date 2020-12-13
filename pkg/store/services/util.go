package services

import "github.com/oars-sigs/oars-cloud/core"

func getKey(selector *core.Service) string {
	prefix := "services/svc/namespaces/"
	if selector.Namespace != "" {
		prefix += selector.Namespace + "/"
	}
	if selector.Name != "" {
		prefix += selector.Name
	}
	return prefix
}

func isContain(src, selector *core.Service) bool {
	if selector.Namespace == "" {
		return true
	}
	if selector.Namespace != "" && selector.Namespace == src.Namespace {
		if selector.Name == "" {
			return true
		}
		if selector.Name != "" && selector.Name == src.Name {
			return true
		}
	}
	return false
}
