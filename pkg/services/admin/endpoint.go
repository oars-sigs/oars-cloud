package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regEndpoint(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetEndPoint(args)
	case "restart":
		return s.RestartEndPoint(args)
	case "stop":
		return s.StopEndPoint(args)
	case "remove":
		return s.RemoveEndPoint(args)
	case "log":
		return s.GetEndPointLog(args)
	case "logstream":
		return s.LogStream(ctx, args)
	case "exec":
		return s.ExecEndPoint(ctx, args)
	case "event":
		return s.GetEndPointEvent(args)
	case "delete":
		return s.DeleteEndPoint(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) getAddr(hostname string) (string, error) {
	r := s.GetEndPoint(core.Endpoint{
		ResourceMeta: &core.ResourceMeta{
			Namespace: "system",
		},
		Service: "node",
	})
	if r.Code != core.ServiceSuccessCode {
		return "", errors.New(r.SubCode)
	}
	res := r.Data.([]core.Resource)
	for _, v := range res {
		ret := v.(*core.Endpoint)
		if ret.Status.ID == hostname {
			return fmt.Sprintf("%s:%d", ret.Status.IP, ret.Status.Port), nil
		}
	}
	return "", errors.New("not found node")

}

func (s *service) GetEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}

	ctx := context.TODO()
	endpoints, err := s.edpStore.List(ctx, &endpoint, &core.ListOptions{})
	return core.NewAPIReply(endpoints)
}

func (s *service) DeleteEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if endpoint.Name == "" {
		return e.InvalidParameterError(err)
	}

	ctx := context.TODO()
	err = s.edpStore.Delete(ctx, &endpoint, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetEndPointEvent(args interface{}) *core.APIReply {
	var edp core.Endpoint
	err := unmarshalArgs(args, &edp)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	event := new(core.Event)
	event.GenName(&edp)
	return s.GetEvent(event)
}

func (s *service) RestartEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	addr, err := s.getAddr(endpoint.Status.Node.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return s.rpcClient.Call(addr, "endpoint.restart", args)
}

func (s *service) StopEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	addr, err := s.getAddr(endpoint.Status.Node.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return s.rpcClient.Call(addr, "endpoint.stop", args)
}

func (s *service) RemoveEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	addr, err := s.getAddr(endpoint.Status.Node.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return s.rpcClient.Call(addr, "endpoint.remove", args)
}

func (s *service) GetEndPointLog(args interface{}) *core.APIReply {
	var opt core.EndpointLogOpt
	err := unmarshalArgs(args, &opt)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	addr, err := s.getAddr(opt.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return s.rpcClient.Call(addr, "endpoint.log", args)
}

func (s *service) LogStream(cc context.Context, args interface{}) *core.APIReply {
	ctx, ok := cc.(*gin.Context)
	if !ok {
		return core.NewAPIError(errors.New("context error"))
	}
	id := ctx.Param("id")
	hostname := ctx.Param("hostname")
	addr, err := s.getAddr(hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	urlstr := fmt.Sprintf("https://%s/clog?id=%s", addr, id)

	resp, err := s.rpcClient.HTTPClient().Get(urlstr)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	defer resp.Body.Close()
	io.Copy(ctx.Writer, resp.Body)
	return core.NewAPIReply("")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *service) ExecEndPoint(cc context.Context, args interface{}) *core.APIReply {
	ctx, ok := cc.(*gin.Context)
	if !ok {
		return core.NewAPIError(errors.New("context error"))
	}
	c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return core.NewAPIError(err)
	}
	defer c.Close()

	hostname := ctx.Param("hostname")
	id := ctx.Param("id")
	col := ctx.Query("col")
	row := ctx.Query("row")

	addr, err := s.getAddr(hostname)
	if err != nil {
		return core.NewAPIError(err)
	}

	addr = fmt.Sprintf("wss://%s/exec?id=%s&col=%s&row=%s", addr, id, col, row)
	err = s.connExec(addr+"&cmd=bash", c)
	if err != nil {
		err = s.connExec(addr+"&cmd=sh", c)
		if err != nil {
			return core.NewAPIError(err)
		}
	}
	return core.NewAPIReply("")
}

func (s *service) connExec(addr string, c *websocket.Conn) error {
	cli, resp, err := s.rpcClient.WSDailer().Dial(addr, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	defer cli.Close()
	stopCh := make(chan error)
	go func() {
		i := 0
		for {
			mt, data, err := cli.ReadMessage()
			if err != nil {
				stopCh <- err
				break
			}
			if i == 0 && strings.HasPrefix(string(data), "OCI runtime exec failed:") {
				stopCh <- errors.New("not found cmd")
				break
			}
			if i == 0 {
				go func() {
					for {
						mt, data, err := c.ReadMessage()
						if err != nil {
							stopCh <- err
							break
						}
						err = cli.WriteMessage(mt, data)
						if err != nil {
							stopCh <- err
							break
						}
					}
				}()
			}
			i = 1
			err = c.WriteMessage(mt, data)
			if err != nil {
				stopCh <- err
				break
			}
			if mt == websocket.CloseMessage {
				stopCh <- nil
				break
			}
		}
	}()
	for {
		select {
		case err := <-stopCh:
			if err != nil {
				return err
			}
			return nil
		}
	}
}
