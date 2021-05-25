package admin

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"

	"github.com/robfig/cron"
)

func (s *service) regCron(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetCron(args)
	case "put":
		return s.PutCron(args)
	case "delete":
		return s.DeleteCron(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) GetCron(args interface{}) *core.APIReply {
	var c core.Cron
	err := unmarshalArgs(args, &c)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	cs, err := s.cronStore.List(ctx, &c, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(cs)
}

func (s *service) PutCron(args interface{}) *core.APIReply {
	var c core.Cron
	err := unmarshalArgs(args, &c)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if !nameRegex.MatchString(c.Name) {
		return e.InvalidParameterError()
	}
	sched, err := cron.Parse(c.Expr)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if c.Status == nil {
		c.Status = new(core.CronStatus)
	}
	c.Status.Next = sched.Next(time.Now()).Unix()
	ctx := context.TODO()
	_, err = s.cronStore.Put(ctx, &c, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(c)
}

func (s *service) DeleteCron(args interface{}) *core.APIReply {
	var c core.Cron
	err := unmarshalArgs(args, &c)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.cronStore.Delete(ctx, &c, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}
