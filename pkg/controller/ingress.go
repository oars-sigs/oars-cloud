package controller

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/oars-sigs/oars-cloud/core"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ingressController struct {
	store      core.KVStore
	cfg        *core.Config
	ecache     sync.Map
	icache     sync.Map
	classCache sync.Map
	snapshot   cachev3.SnapshotCache
}

func (c *ingressController) run(stopCh <-chan struct{}) error {
	cache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, log.New())
	c.snapshot = cache
	ctx := context.Background()
	cb := &testv3.Callbacks{Debug: false}
	srv := serverv3.NewServer(ctx, cache, cb)
	runServer(ctx, srv, c.cfg.Ingress.XDSPort)
	return nil
}

func (c *ingressController) watchIngressListener(stopCh <-chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "ingresses/listener/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true})
	for {
		select {
		case res := <-updateCh:
			kv := res.KV
			if !res.Put {
				kv = res.PrevKV
			}
			class := new(core.IngressListener)
			err := class.Parse(kv.Value)
			if err != nil {
				log.Error(err)
				continue
			}
			if res.Put {
				c.classCache.Store(class.Name, class)
				continue
			}
			c.classCache.Delete(class.Name)

		case err := <-errCh:
			fmt.Println(err)
		case <-stopCh:
			cancel()
			return nil
		}
	}
}

func (c *ingressController) watchIngress(stopCh <-chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "ingresses/instance/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true})
	for {
		select {
		case res := <-updateCh:
			kv := res.KV
			if !res.Put {
				kv = res.PrevKV
			}
			ingress := new(core.Ingress)
			err := ingress.Parse(kv.Value)
			if err != nil {
				log.Error(err)
				continue
			}
			if res.Put {
				c.icache.Store(ingress.Name, ingress)
				continue
			}
			c.icache.Delete(ingress.Name)

		case err := <-errCh:
			fmt.Println(err)
		case <-stopCh:
			cancel()
			return nil
		}
	}
}

func (c *ingressController) watchEndpoint(stopCh <-chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "services/endpoint/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true})
	for {
		select {
		case res := <-updateCh:
			kv := res.KV
			if !res.Put {
				kv = res.PrevKV
			}
			endpoint := new(core.Endpoint)
			err := endpoint.Parse(kv.Value)
			if err != nil {
				log.Error(err)
				continue
			}
			if endpoint.Service == "node" {
				continue
			}
			if res.Put {
				olde, ok := c.ecache.LoadOrStore(endpoint.ID, endpoint)
				if ok {
					if olde.(*core.Endpoint).HostIP == endpoint.HostIP {
						continue
					}
				}
				if !ok {
					if olde.(*core.Endpoint).State != "running" {
						continue
					}
				}
				c.updateHandle(endpoint)
				continue
			}
			c.updateHandle(endpoint)
			c.ecache.Delete(endpoint.ID)
		case err := <-errCh:
			fmt.Println(err)
		case <-stopCh:
			cancel()
			return nil
		}
	}
}

func (c *ingressController) updateHandle(endpoint *core.Endpoint) {
	clusters := make([]*cluster.Cluster, 0)
	endpoints := make(map[string][]string)
	c.ecache.Range(func(k, v interface{}) bool {
		endpoint := v.(*core.Endpoint)
		if _, ok := endpoints[endpoint.Name]; !ok {
			endpoints[endpoint.Name] = make([]string, 0)
		}
		endpoints[endpoint.Name] = append(endpoints[endpoint.Name], endpoint.HostIP)
		return true
	})
	router := &route.RouteConfiguration{
		Name:         "oars_ingress",
		VirtualHosts: make([]*route.VirtualHost, 0),
	}
	c.icache.Range(func(k, v interface{}) bool {
		ingress := v.(*core.Ingress)
		for _, rule := range ingress.Rules {
			virtualHost := &route.VirtualHost{
				Name:    c.getHostName(rule.Host),
				Domains: []string{rule.Host},
				Routes:  make([]*route.Route, 0),
			}

			for _, path := range rule.HTTP.Paths {
				clusterName := fmt.Sprintf("%s_%s_%d", path.Backend.Namespace, path.Backend.ServiceName, path.Backend.ServicePort)
				r := &route.Route{
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: path.Path,
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: clusterName,
							},
						},
					},
				}
				virtualHost.Routes = append(virtualHost.Routes, r)
				c := &cluster.Cluster{
					Name: clusterName,
				}
				clusters = append(clusters, c)
			}
			router.VirtualHosts = append(router.VirtualHosts, virtualHost)
		}
		return true
	})

}

func (c *ingressController) getHostName(host string) string {
	if host == "" {
		host = "all"
	}
	return "h_" + strings.Replace(host, ".", "_", -1)
}

const (
	grpcMaxConcurrentStreams = 1000000
)

func registerServer(grpcServer *grpc.Server, server serverv3.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}

// RunServer starts an xDS server at the given port.
func runServer(ctx context.Context, srv3 serverv3.Server, port int) error {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error(err)
		return err
	}
	registerServer(grpcServer, srv3)
	log.Info("management server listening on %d\n", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
