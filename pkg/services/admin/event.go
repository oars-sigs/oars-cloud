package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regEvent(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetEvent(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) GetEvent(args interface{}) *core.APIReply {
	var event core.Event
	err := unmarshalArgs(args, &event)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	events, err := s.eventStore.List(ctx, &event, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(events)
}
