package namespaces

import "github.com/oars-sigs/oars-cloud/core"

func getKey(selector *core.Namespace) string {
	prefix := "namespaces/"
	return prefix + selector.Name
}
