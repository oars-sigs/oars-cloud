package controller

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/oars-sigs/oars-cloud/core"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ingressController struct {
	store         core.KVStore
	cfg           *core.Config
	ecache        sync.Map
	icache        sync.Map
	listenerCache sync.Map
	snapshot      cachev3.SnapshotCache
	version       int
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
			lis := new(core.IngressListener)
			err := lis.Parse(kv.Value)
			if err != nil {
				log.Error(err)
				continue
			}
			if res.Put {
				c.listenerCache.Store(lis.Name, lis)
				continue
			}
			c.listenerCache.Delete(lis.Name)

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
	go c.store.Watch(ctx, "ingresses/route/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true})
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
	clusters := make([]types.Resource, 0)
	endpoints := make(map[string][]string)
	c.ecache.Range(func(k, v interface{}) bool {
		endpoint := v.(*core.Endpoint)
		key := endpoint.Namespace + "_" + endpoint.Service
		if _, ok := endpoints[key]; !ok {
			endpoints[key] = make([]string, 0)
		}
		endpoints[key] = append(endpoints[key], endpoint.HostIP)
		return true
	})
	routers := make(map[string]*route.RouteConfiguration)
	filterChains := make(map[string][]*listener.FilterChain)
	c.icache.Range(func(k, v interface{}) bool {
		ingress := v.(*core.Ingress)
		//route
		for _, rule := range ingress.Rules {
			virtualHost := &route.VirtualHost{
				Name:    c.getHostName(rule.Host),
				Domains: []string{rule.Host},
				Routes:  make([]*route.Route, 0),
			}

			for _, path := range rule.HTTP.Paths {
				clusterName := fmt.Sprintf("%s_%s_%d", ingress.Namespace, path.Backend.ServiceName, path.Backend.ServicePort)
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

				key := ingress.Namespace + "_" + path.Backend.ServiceName
				eps, ok := endpoints[key]
				if !ok {
					continue
				}
				cla := &endpointv3.ClusterLoadAssignment{
					ClusterName: clusterName,
					Endpoints:   makeEndpoints(eps, path.Backend.ServicePort),
				}
				c := &cluster.Cluster{
					Name:                 clusterName,
					ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
					ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
					LbPolicy:             cluster.Cluster_ROUND_ROBIN,
					LoadAssignment:       cla,
					DnsLookupFamily:      cluster.Cluster_V4_ONLY,
				}
				clusters = append(clusters, c)
				virtualHost.Routes = append(virtualHost.Routes, r)
			}
			name := ingress.Namespace + "_" + ingress.Name
			_, ok := routers[name]
			if !ok {
				routers[name] = &route.RouteConfiguration{
					Name:         name,
					VirtualHosts: make([]*route.VirtualHost, 0),
				}
			}
			routers[name].VirtualHosts = append(routers[name].VirtualHosts, virtualHost)
			//filterChains
			manager := &hcm.HttpConnectionManager{
				CodecType:  hcm.HttpConnectionManager_AUTO,
				StatPrefix: "http",
				RouteSpecifier: &hcm.HttpConnectionManager_Rds{
					Rds: &hcm.Rds{
						ConfigSource:    makeConfigSource(),
						RouteConfigName: name,
					},
				},
				HttpFilters: []*hcm.HttpFilter{{
					Name: wellknown.Router,
				}},
			}
			pbst, err := ptypes.MarshalAny(manager)
			if err != nil {
				return true
			}
			if _, ok := filterChains[ingress.Listener]; !ok {
				filterChains[ingress.Listener] = make([]*listener.FilterChain, 0)
			}
			filterChain := &listener.FilterChain{
				Filters: []*listener.Filter{{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				}},
			}
			filterChains[ingress.Listener] = append(filterChains[ingress.Listener], filterChain)
		}
		return true
	})
	listeners := make([]types.Resource, 0)
	c.listenerCache.Range(func(k, v interface{}) bool {
		lis := v.(*core.IngressListener)
		if _, ok := filterChains[lis.Name]; !ok {
			filterChains[lis.Name] = make([]*listener.FilterChain, 0)
		}
		liser := &listener.Listener{
			Name: lis.Name,
			Address: &corev3.Address{
				Address: &corev3.Address_SocketAddress{
					SocketAddress: &corev3.SocketAddress{
						Protocol: corev3.SocketAddress_TCP,
						Address:  "0.0.0.0",
						PortSpecifier: &corev3.SocketAddress_PortValue{
							PortValue: uint32(lis.Port),
						},
					},
				},
			},
			FilterChains: filterChains[lis.Name],
		}
		listeners = append(listeners, liser)
		return true
	})
	c.version++
	rResources := make([]types.Resource, 0)
	for k := range routers {
		rResources = append(rResources, routers[k])
	}
	snap := cachev3.NewSnapshot(
		fmt.Sprintf("v.%d", c.version),
		[]types.Resource{}, //endpoints
		clusters,           //clusters
		rResources,         //router
		listeners,          //listeners
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)
	c.snapshot.SetSnapshot("oars_ingress", snap)

}

func (c *ingressController) getHostName(host string) string {
	if host == "" {
		host = "all"
	}
	return "h_" + strings.Replace(host, ".", "_", -1)
}

func makeEndpoints(ips []string, port int) []*endpointv3.LocalityLbEndpoints {
	lbes := make([]*endpointv3.LocalityLbEndpoints, 0)
	for _, ip := range ips {
		lbes = append(lbes,
			&endpointv3.LocalityLbEndpoints{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Protocol: corev3.SocketAddress_TCP,
										Address:  ip,
										PortSpecifier: &corev3.SocketAddress_PortValue{
											PortValue: uint32(port),
										},
									},
								},
							},
						},
					},
				}},
			},
		)
	}
	return lbes
}

const (
	grpcMaxConcurrentStreams = 1000000
)

func makeConfigSource() *corev3.ConfigSource {
	source := &corev3.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &corev3.ConfigSource_ApiConfigSource{
		ApiConfigSource: &corev3.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   corev3.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*corev3.GrpcService{{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}

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
	log.Infof("management server listening on %d", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
