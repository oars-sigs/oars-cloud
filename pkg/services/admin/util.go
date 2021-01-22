package admin

import (
	"context"

	"github.com/ghodss/yaml"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regUtil(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "yamlFormat":
		return s.FormatYaml(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) FormatYaml(args interface{}) *core.APIReply {
	data, err := yaml.Marshal(args)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return core.NewAPIReply(string(data))
}
