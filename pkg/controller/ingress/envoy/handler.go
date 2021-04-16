package envoy

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sort"
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
	tcpproxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
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
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ingress struct {
	snapshot       cachev3.SnapshotCache
	icache         sync.Map
	listenerCache  sync.Map
	listenerLister core.ResourceLister
	routeLister    core.ResourceLister
	certLister     core.ResourceLister
	version        int64
}

func New(listenerLister, routeLister, certLister core.ResourceLister, port int) core.IngressControllerHandle {
	snapshot := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, log.New())
	cb := &testv3.Callbacks{Debug: false}
	ctx := context.Background()
	srv := serverv3.NewServer(ctx, snapshot, cb)
	runServer(ctx, srv, port)
	return &ingress{
		snapshot:       snapshot,
		listenerLister: listenerLister,
		routeLister:    routeLister,
		certLister:     certLister,
	}
}

type ingressRule struct {
	namespace string
	core.IngressRule
}

func (c *ingress) UpdateHandle() {
	rules := make(map[string]map[string][]ingressRule)
	routeList, _ := c.routeLister.List()
	listenerList, _ := c.listenerLister.List()
	for _, v := range routeList {
		ingress := v.(*core.IngressRoute)
		if _, ok := rules[ingress.Listener]; !ok {
			rules[ingress.Listener] = make(map[string][]ingressRule)
		}
		for _, rule := range ingress.Rules {
			if _, ok := rules[ingress.Listener][rule.Host]; !ok {
				rules[ingress.Listener][rule.Host] = make([]ingressRule, 0)
			}
			ir := ingressRule{
				namespace:   ingress.Namespace,
				IngressRule: rule,
			}
			rules[ingress.Listener][rule.Host] = append(rules[ingress.Listener][rule.Host], ir)
		}
	}
	listeners := make([]types.Resource, 0)
	routers := make([]types.Resource, 0)
	clustersMap := make(map[string]*cluster.Cluster)
	for _, v := range listenerList {
		lis := v.(*core.IngressListener)
		if lis.Drive != "" && lis.Drive != "envoy" {
			continue
		}
		if _, ok := rules[lis.Name]; !ok {
			continue
		}
		filterChains, newRouters := c.makeTCPChains(lis, rules[lis.Name], clustersMap)
		if len(filterChains) == 0 {
			filterChains, newRouters = c.makeHTTPChains(lis, rules[lis.Name], clustersMap)
		}
		routers = append(routers, newRouters...)
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
	}
	c.version = time.Now().UnixNano()
	clusters := make([]types.Resource, 0)
	for k := range clustersMap {
		clusters = append(clusters, clustersMap[k])
	}
	//dd, _ := json.Marshal(listeners)
	//fmt.Println(string(dd))
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

func (c *ingress) makeTCPChains(lis *core.IngressListener, rules map[string][]ingressRule, clustersMap map[string]*cluster.Cluster) ([]*listener.FilterChain, []types.Resource) {
	filterChains := make([]*listener.FilterChain, 0)
	routers := make([]types.Resource, 0)
	for _, irs := range rules {
		for _, ir := range irs {
			if ir.TCP == nil {
				continue
			}
			clusterName := c.getClusterName(ir.TCP.Backend.ServiceName, ir.namespace, ir.TCP.Backend.ServicePort)
			if _, ok := clustersMap[clusterName]; !ok {
				cla := &endpointv3.ClusterLoadAssignment{
					ClusterName: clusterName,
					Endpoints:   makeEndpoints([]string{ir.TCP.Backend.ServiceName + "." + ir.namespace}, ir.TCP.Backend.ServicePort),
				}
				clustersMap[clusterName] = &cluster.Cluster{
					Name:                 clusterName,
					ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
					ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
					LbPolicy:             cluster.Cluster_ROUND_ROBIN,
					LoadAssignment:       cla,
					DnsLookupFamily:      cluster.Cluster_V4_ONLY,
				}
			}
			tcpp := &tcpproxy.TcpProxy{
				StatPrefix: "ingress_tcp",
				ClusterSpecifier: &tcpproxy.TcpProxy_Cluster{
					Cluster: clusterName,
				},
			}
			pbst, err := ptypes.MarshalAny(tcpp)
			if err != nil {
				continue
			}
			filterChain := &listener.FilterChain{
				Filters: []*listener.Filter{{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				}},
			}
			filterChains = append(filterChains, filterChain)
		}
	}
	return filterChains, routers
}

func (c *ingress) getClusterName(svc, ns string, port int) string {
	return fmt.Sprintf("%s_%s_%d", svc, ns, port)
}

func (c *ingress) makeHTTPChains(lis *core.IngressListener, rules map[string][]ingressRule, clustersMap map[string]*cluster.Cluster) ([]*listener.FilterChain, []types.Resource) {
	filterChains := make([]*listener.FilterChain, 0)
	routers := make([]types.Resource, 0)
	virtualHosts := make([]*route.VirtualHost, 0)
	//range and create one listener all hosts
	for host, irs := range rules {
		routes := make([]*route.Route, 0)

		//range and create one host all routes
		for _, ir := range irs {
			sort.Slice(ir.HTTP.Paths, func(i, j int) bool {
				return strings.HasPrefix(ir.HTTP.Paths[i].Path, ir.HTTP.Paths[j].Path)
			})
			fmt.Println("************")
			for _, path := range ir.HTTP.Paths {
				fmt.Println(path.Path)
				//generate one route in a host
				clusterName := c.getClusterName(path.Backend.ServiceName, ir.namespace, path.Backend.ServicePort)
				r := &route.Route{
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: path.Path,
						},
					},
				}
				rr := &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				}
				for k, v := range path.Config {
					switch k {
					case "prefix_rewrite":
						rr.Route.PrefixRewrite = v
					case "auto_host_rewrite":
						rr.Route.HostRewriteSpecifier = &route.RouteAction_AutoHostRewrite{
							AutoHostRewrite: &wrappers.BoolValue{Value: true},
						}
					case "host_rewrite":
						rr.Route.HostRewriteSpecifier = &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: v,
						}
					}
				}
				r.Action = rr
				routes = append(routes, r)

				//generate cluster if not exist
				if _, ok := clustersMap[clusterName]; !ok {
					cla := &endpointv3.ClusterLoadAssignment{
						ClusterName: clusterName,
						Endpoints:   makeEndpoints([]string{path.Backend.ServiceName + "." + ir.namespace}, path.Backend.ServicePort),
					}
					clustersMap[clusterName] = &cluster.Cluster{
						Name:                 clusterName,
						ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
						ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
						LbPolicy:             cluster.Cluster_ROUND_ROBIN,
						LoadAssignment:       cla,
						DnsLookupFamily:      cluster.Cluster_V4_ONLY,
					}
				}

			}
		}

		//router
		routeName := fmt.Sprintf("%s_%s", lis.Name, strings.ReplaceAll(host, ".", "_"))
		//add http virtualHosts
		if lis.DisabledTLS {
			virtualHosts = append(virtualHosts, &route.VirtualHost{
				Name:    routeName,
				Domains: []string{host},
				Routes:  routes,
			})
			continue
		}
		//add  https router
		router := &route.RouteConfiguration{
			Name: routeName,
			VirtualHosts: []*route.VirtualHost{
				&route.VirtualHost{
					Name:    "default",
					Domains: []string{"*"},
					Routes:  routes,
				},
			},
		}
		routers = append(routers, router)

		//hcm
		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "ingress_http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: routeName,
				},
			},
			HttpFilters: []*hcm.HttpFilter{{
				Name: wellknown.Router,
			}},
			UpgradeConfigs: []*hcm.HttpConnectionManager_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
			},
		}
		pbst, err := ptypes.MarshalAny(manager)
		if err != nil {
			continue
		}

		// generate tls cert
		cert, key := c.getCert(host, lis)
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
			continue
		}

		filterChain := &listener.FilterChain{
			FilterChainMatch: &listener.FilterChainMatch{
				ServerNames: []string{host},
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
	//add http route
	if len(virtualHosts) > 0 {
		router := &route.RouteConfiguration{
			Name:         "http_without_tls",
			VirtualHosts: virtualHosts,
		}
		routers = append(routers, router)
		//hcm
		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "ingress_http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: "http_without_tls",
				},
			},
			HttpFilters: []*hcm.HttpFilter{{
				Name: wellknown.Router,
			}},
			UpgradeConfigs: []*hcm.HttpConnectionManager_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
			},
		}
		pbst, err := ptypes.MarshalAny(manager)
		if err != nil {

		}
		filterChain := &listener.FilterChain{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}
		filterChains = append(filterChains, filterChain)
	}
	return filterChains, routers
}

func (c *ingress) getCert(host string, lis *core.IngressListener) ([]byte, []byte) {
	score := 0
	index := -1
	certRes, _ := c.certLister.List()
	for i, tlsCert := range certRes {
		cert := tlsCert.(*core.Certificate)
		if cert.Cert == "" || cert.Info.IsCA || len(cert.Info.Domains) == 0 {
			continue
		}
		cerths := strings.Split(cert.Info.Domains[0], ".")
		hs := strings.Split(host, ".")
		if len(cerths) > len(hs) {
			continue
		}
		curScore := 0
		for n := len(cerths) - 1; n >= 0; n-- {
			ch := cerths[n]
			if hs[n] == ch {
				curScore += 2
				continue
			}
			if ch == "*" {
				curScore++
				break
			}
			if hs[n] != ch {
				curScore = 0
				break
			}
		}
		if curScore > score {
			index = i
			score = curScore
		}
	}
	if index != -1 {
		cert, _ := base64.StdEncoding.DecodeString(certRes[index].(*core.Certificate).Cert)
		key, _ := base64.StdEncoding.DecodeString(certRes[index].(*core.Certificate).Key)
		return cert, key
	}
	return nil, nil
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
