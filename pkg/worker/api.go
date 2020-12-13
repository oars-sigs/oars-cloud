package worker

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"

	"github.com/gorilla/websocket"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/sirupsen/logrus"
)

type rpcServer struct {
	d *daemon
}

func (s *rpcServer) EndpointRestart(endpoint *core.Endpoint, reply *core.APIReply) error {
	ctx := context.Background()
	return s.d.Restart(ctx, endpoint.Status.ID)
}

func (s *rpcServer) EndpointStop(endpoint *core.Endpoint, reply *core.APIReply) error {
	return s.d.Stop(context.Background(), endpoint.Status.ID)
}

func (s *rpcServer) EndpointLog(opt *core.EndpointLogOpt, reply *core.APIReply) error {
	ctx := context.Background()
	l, err := s.d.Log(ctx, opt.ID, opt.Tail, opt.Since)
	if err != nil {
		return err
	}
	*reply = *core.NewAPIReply(l)
	return nil
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
			_, err = io.Copy(resp.Conn, reader)
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
			n, err := resp.Reader.Read(buf)
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
	err := rpc.RegisterName("Endpoint", &rpcServer{d})
	if err != nil {
		fmt.Println(err)
	}
	rpc.HandleHTTP()
	http.HandleFunc("/exec", d.exec)
	fmt.Printf("Start RPC server in :%d\n", d.node.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", d.node.Port))
	if err != nil {
		return err
	}

	err = http.Serve(lis, nil)
	return err
}
