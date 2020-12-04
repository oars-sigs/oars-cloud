package controller

import (
	"context"
	"encoding/base64"
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
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ingressController struct {
	store         core.KVStore
	cfg           *core.Config
	snapshot      cachev3.SnapshotCache
	ecache        sync.Map
	icache        sync.Map
	listenerCache sync.Map
	useServices   sync.Map
	version       int
}

func newIngress(store core.KVStore, cfg *core.Config) *ingressController {
	snapshot := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, log.New())
	return &ingressController{store: store, cfg: cfg, snapshot: snapshot}
}

func (c *ingressController) run(stopCh <-chan struct{}) error {
	ctx := context.Background()
	cb := &testv3.Callbacks{Debug: false}
	srv := serverv3.NewServer(ctx, c.snapshot, cb)
	lrev, irev, erev, err := c.initCache()
	if err != nil {
		return err
	}
	c.updateHandle(nil)
	go c.watchIngressListener(lrev, stopCh)
	go c.watchIngress(irev, stopCh)
	go c.watchEndpoint(erev, stopCh)
	runServer(ctx, srv, c.cfg.Ingress.XDSPort)
	return nil
}
func (c *ingressController) initCache() (lrev int64, irev int64, erev int64, err error) {
	var kvs []core.KV
	kvs, lrev, err = c.store.GetWithRev(context.Background(), "ingresses/listener/", core.KVOption{WithPrefix: true})
	if err != nil {
		return
	}
	for _, kv := range kvs {
		r := new(core.IngressListener)
		err = r.Parse(kv.Value)
		if err != nil {
			return
		}
		c.listenerCache.Store(r.Name, r)
	}
	kvs, irev, err = c.store.GetWithRev(context.Background(), "ingresses/route/", core.KVOption{WithPrefix: true})
	if err != nil {
		return
	}
	for _, kv := range kvs {
		r := new(core.Ingress)
		err = r.Parse(kv.Value)
		if err != nil {
			return
		}
		c.icache.Store(r.Name, r)
	}
	kvs, erev, err = c.store.GetWithRev(context.Background(), "ingresses/endpoint/", core.KVOption{WithPrefix: true})
	if err != nil {
		return
	}
	for _, kv := range kvs {
		r := new(core.Endpoint)
		err = r.Parse(kv.Value)
		if err != nil {
			return
		}
		c.ecache.Store(r.ID, r)
	}

	return

}

func (c *ingressController) watchIngressListener(rev int64, stopCh <-chan struct{}) error {

	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "ingresses/listener/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true, DisableFirst: true, WithRev: rev})
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

func (c *ingressController) watchIngress(rev int64, stopCh <-chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "ingresses/route/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true, DisableFirst: true, WithRev: rev})
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

func (c *ingressController) watchEndpoint(rev int64, stopCh <-chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "services/endpoint/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true, DisableFirst: true, WithRev: rev})
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
	if endpoint != nil {
		if _, ok := c.useServices.Load(endpoint.Service); !ok {
			return
		}
	}
	c.useServices = sync.Map{}
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
	type ingressRs struct {
		namespace string
		rules     []core.IngressRule
	}
	rules := make(map[string]map[string]*ingressRs)
	c.icache.Range(func(k, v interface{}) bool {
		ingress := v.(*core.Ingress)
		if _, ok := rules[ingress.Listener]; !ok {
			rules[ingress.Listener] = make(map[string]*ingressRs)
		}
		for _, rule := range ingress.Rules {
			if _, ok := rules[ingress.Listener][rule.Host]; !ok {
				rules[ingress.Listener][rule.Host] = &ingressRs{
					namespace: ingress.Namespace,
					rules:     make([]core.IngressRule, 0),
				}
			}
			rules[ingress.Listener][rule.Host].rules = append(rules[ingress.Listener][rule.Host].rules, rule)
		}
		return true
	})
	listeners := make([]types.Resource, 0)
	routers := make([]types.Resource, 0)
	c.listenerCache.Range(func(k, v interface{}) bool {
		lis := v.(*core.IngressListener)
		filterChains := make([]*listener.FilterChain, 0)
		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: lis.Name,
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
		if _, ok := rules[lis.Name]; ok {
			for _, rs := range rules[lis.Name] {
				router := &route.RouteConfiguration{
					Name:         lis.Name,
					VirtualHosts: make([]*route.VirtualHost, 0),
				}
				for _, r := range rs.rules {
					virtualHost := &route.VirtualHost{
						Name:    c.getHostName(r.Host),
						Domains: []string{r.Host},
						Routes:  make([]*route.Route, 0),
					}

					for _, path := range r.HTTP.Paths {
						c.useServices.Store(path.Backend.ServicePort, struct{}{})
						clusterName := fmt.Sprintf("r_%s_%s", strings.ReplaceAll(r.Host, ".", "_"), strings.ReplaceAll(path.Path, "/", "_"))
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

						key := rs.namespace + "_" + path.Backend.ServiceName
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

					router.VirtualHosts = append(router.VirtualHosts, virtualHost)
					cert, key := c.getCert(r.Host, lis)
					tls := &tlsv3.DownstreamTlsContext{
						CommonTlsContext: &tlsv3.CommonTlsContext{
							TlsCertificates: []*tlsv3.TlsCertificate{
								&tlsv3.TlsCertificate{
									PrivateKey: &corev3.DataSource{
										Specifier: &corev3.DataSource_InlineBytes{
											InlineBytes: key,
										},
									},
									CertificateChain: &corev3.DataSource{
										Specifier: &corev3.DataSource_InlineBytes{
											InlineBytes: cert,
										},
									},
								},
							},
							//TODO: 使用sds
							// TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
							// 	&tlsv3.SdsSecretConfig{
							// 		SdsConfig: makeConfigSource(),
							// 	},
							// },
						},
					}
					pbtls, err := ptypes.MarshalAny(tls)
					if err != nil {
						return true
					}
					filterChain := &listener.FilterChain{
						FilterChainMatch: &listener.FilterChainMatch{
							ServerNames: []string{r.Host},
						},
						TransportSocket: &corev3.TransportSocket{
							Name: "envoy.transport_sockets.tls",
							ConfigType: &corev3.TransportSocket_TypedConfig{
								TypedConfig: pbtls,
							},
						},
						Filters: []*listener.Filter{{
							Name: wellknown.HTTPConnectionManager,
							ConfigType: &listener.Filter_TypedConfig{
								TypedConfig: pbst,
							},
						}},
					}

					filterChains = append(filterChains, filterChain)
				}
				routers = append(routers, router)
			}
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
			FilterChains: filterChains,
		}
		listeners = append(listeners, liser)
		return true
	})
	c.version++

	snap := cachev3.NewSnapshot(
		fmt.Sprintf("v.%d", c.version),
		[]types.Resource{}, //endpoints
		clusters,           //clusters
		routers,            //routers
		listeners,          //listeners
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)
	c.snapshot.SetSnapshot("oars_ingress", snap)

}

func (c *ingressController) getCert(host string, lis *core.IngressListener) ([]byte, []byte) {
	if lis.TLSCerts != nil {
		tlsCert, ok := lis.TLSCerts[host]
		if !ok {
			tlsCert, ok = lis.TLSCerts["*"]
			if !ok {
				return nil, nil
			}
		}
		cert, _ := base64.StdEncoding.DecodeString(tlsCert.Cert)
		key, _ := base64.StdEncoding.DecodeString(tlsCert.Key)
		return cert, key
	}
	return nil, nil
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
	source.ResourceApiVersion = corev3.ApiVersion_V3
	source.ConfigSourceSpecifier = &corev3.ConfigSource_ApiConfigSource{
		ApiConfigSource: &corev3.ApiConfigSource{
			TransportApiVersion:       corev3.ApiVersion_V3,
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
