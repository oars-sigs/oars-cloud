package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/rpc"
	"github.com/sirupsen/logrus"
)

type rpcServer struct {
	d *daemon
}

func (s *rpcServer) EndpointRestart(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := rpc.UnmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.Background()
	err = s.d.Restart(ctx, endpoint.Status.ID)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(nil)
}

func (s *rpcServer) EndpointStop(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := rpc.UnmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	err = s.d.Stop(context.Background(), endpoint.Status.ID)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(nil)
}

func (s *rpcServer) EndpointLog(args interface{}) *core.APIReply {
	var opt core.EndpointLogOpt
	err := rpc.UnmarshalArgs(args, &opt)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.Background()
	l, err := s.d.Log(ctx, opt.ID, opt.Tail, opt.Since)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(l)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (d *daemon) exec(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	cmd := r.FormValue("cmd")
	if cmd == "" {
		cmd = "sh"
	}
	resp, err := d.Exec(id, cmd)
	if err != nil {
		logrus.Error(err)
		return
	}

	defer resp.Close()
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer c.Close()
	stopCh := make(chan struct{})
	go func() {
		for {
			_, reader, err := c.NextReader()
			if err != nil {
				logrus.Error(err)
				stopCh <- struct{}{}
				break
			}
			_, err = io.Copy(resp, reader)
			if err != nil {
				logrus.Error(err)
				stopCh <- struct{}{}
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := resp.Read(buf)
			if err != err {
				logrus.Error(err)
				stopCh <- struct{}{}
				break
			}
			err = c.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				logrus.Error(err)
				stopCh <- struct{}{}
				break
			}
		}
	}()
	<-stopCh
}

func (d *daemon) reg() error {
	api := &rpcServer{d}
	d.rpcServer.Register("endpoint.stop", api.EndpointStop)
	d.rpcServer.Register("endpoint.restart", api.EndpointRestart)
	d.rpcServer.Register("endpoint.log", api.EndpointLog)
	http.HandleFunc("/exec", d.exec)
	fmt.Printf("Start RPC server in :%d\n", d.node.Port)
	return d.rpcServer.Listen()
}
