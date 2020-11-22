package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/oars-sigs/oars-cloud/core"
)

type rpcServer struct {
	d *daemon
}

func (s *rpcServer) EndpointRestart(endpoint *core.Endpoint, reply *core.APIReply) error {
	ctx := context.Background()
	return s.d.Restart(ctx, endpoint.ID)
}

func (s *rpcServer) EndpointStop(endpoint *core.Endpoint, reply *core.APIReply) error {
	return s.d.Stop(context.Background(), endpoint.ID)
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

func (d *daemon) reg() error {
	err := rpc.RegisterName("Endpoint", &rpcServer{d})
	if err != nil {
		fmt.Println(err)
	}
	rpc.HandleHTTP()
	fmt.Printf("Start RPC server in :%d\n", d.node.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", d.node.Port))
	if err != nil {
		return err
	}

	err = http.Serve(lis, nil)
	return err
}
