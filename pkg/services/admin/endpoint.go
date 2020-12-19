package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
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
	case "log":
		return s.GetEndPointLog(args)
	case "exec":
		return s.ExecEndPoint(ctx, args)
	}
	return e.MethodNotFoundMethod()
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

func (s *service) RestartEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(endpoint.Status.Node.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	defer conn.Close()
	err = conn.Call("Endpoint.EndpointRestart", endpoint, nil)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return core.NewAPIReply("")
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
	res := r.Data.([]*core.Endpoint)
	for _, ret := range res {
		if ret.Status.ID == hostname {
			return fmt.Sprintf("%s:%d", ret.Status.IP, ret.Status.Port), nil
		}
	}
	return "", errors.New("not found node")

}
func (s *service) getConn(hostname string) (*rpc.Client, error) {
	addr, err := s.getAddr(hostname)
	if err != nil {
		return nil, err
	}
	return rpc.DialHTTP("tcp", addr)

}

func (s *service) StopEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(endpoint.Status.Node.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	defer conn.Close()
	err = conn.Call("Endpoint.EndpointStop", &endpoint, nil)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetEndPointLog(args interface{}) *core.APIReply {
	var opt core.EndpointLogOpt
	err := unmarshalArgs(args, &opt)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(opt.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	defer conn.Close()
	var r core.APIReply
	err = conn.Call("Endpoint.EndpointLog", &opt, &r)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return &r
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

	addr, err := s.getAddr(hostname)
	if err != nil {
		return core.NewAPIError(err)
	}

	addr = "ws://" + addr + "/exec?id=" + id
	err = s.connExec(addr+"&cmd=bash", c)
	if err != nil {
		err = s.connExec(addr+"&cmd=sh", c)
		if err != nil {
			return core.NewAPIError(err)
		}
	}
	fmt.Println(addr)
	return core.NewAPIReply("")
}

func (s *service) connExec(addr string, c *websocket.Conn) error {
	cli, resp, err := websocket.DefaultDialer.Dial(addr, nil)
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
