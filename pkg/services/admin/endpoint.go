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

func (s *service) GetEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	key := fmt.Sprintf("services/endpoint/namespaces/%s/%s", endpoint.Namespace, endpoint.Service)
	if endpoint.Service != "" {
		key += "/"
	}
	if endpoint.Hostname != "" {
		key += endpoint.Hostname
	}
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	endpoints := make([]core.Endpoint, 0)
	for _, kv := range kvs {
		endpoint := new(core.Endpoint)
		endpoint.Parse(kv.Value)
		endpoints = append(endpoints, *endpoint)
	}
	return core.NewAPIReply(endpoints)
}

func (s *service) RestartEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(endpoint.Hostname)
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
		Namespace: "system",
		Service:   "node",
		Hostname:  hostname,
	})
	if r.Code != core.ServiceSuccessCode {
		return "", errors.New(r.SubCode)
	}
	res := r.Data.([]core.Endpoint)
	if len(res) == 0 {
		return "", errors.New("not found node")
	}
	return fmt.Sprintf("%s:%d", res[0].HostIP, res[0].Port), nil
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
	conn, err := s.getConn(endpoint.Hostname)
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
