package traefik

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/oars-sigs/oars-cloud/core"
)

type ingress struct {
	listenerLister core.ResourceLister
	routeLister    core.ResourceLister
	certLister     core.ResourceLister
	data           []byte
	mu             *sync.Mutex
}

func New(listenerLister, routeLister, certLister core.ResourceLister, port int) core.IngressControllerHandle {
	h := &ingress{
		listenerLister: listenerLister,
		routeLister:    routeLister,
		certLister:     certLister,
		mu:             new(sync.Mutex),
	}
	go h.server(port)
	return h
}

func (c *ingress) UpdateHandle() {
	routeList, ok := c.routeLister.List()
	if !ok {
		return
	}
	listenerList, _ := c.listenerLister.List()
	svcs := make(map[string]service)
	routers := make(map[string]router, 0)
	for _, v := range routeList {
		ingress := v.(*core.IngressRoute)
		disTLS := false
		filter := false
		for _, v := range listenerList {
			lis := v.(*core.IngressListener)
			if lis.Name == ingress.Listener {
				disTLS = lis.DisabledTLS
				if lis.Drive == "traefik" {
					filter = true
				}
			}
		}
		if !filter {
			continue
		}

		for _, rule := range ingress.Rules {
			if rule.HTTP != nil {
				host := ""
				if rule.Host != "" {
					host = fmt.Sprintf("Host(`%s`) &&", rule.Host)
				}

				for _, path := range rule.HTTP.Paths {
					svcName := getServiceName(path.Backend.ServiceName, ingress.Namespace, path.Backend.ServicePort) + "_http"
					r := router{
						Rule:        fmt.Sprintf("%s PathPrefix(`%s`)", host, path.Path),
						EntryPoints: []string{ingress.Listener},
						Service:     svcName,
					}
					if !disTLS {
						r.TLS = &routerTLS{
							Domains: []routerTLSDomain{
								routerTLSDomain{
									Main: rule.Host,
									Sans: []string{"*." + strings.Join(strings.Split(rule.Host, ".")[1:], ".")},
								},
							},
						}
					}
					rn := fmt.Sprintf("%s_%s_%s_%s", ingress.Name, ingress.Namespace, ingress.Listener, base64.StdEncoding.EncodeToString([]byte(path.Path)))
					routers[rn] = r
					if _, ok := svcs[svcName]; !ok {
						svcs[svcName] = service{
							LoadBalancer: serviceLB{
								Servers: []serviceServer{
									serviceServer{
										URL: fmt.Sprintf("http://%s.%s:%d", path.Backend.ServiceName, ingress.Namespace, path.Backend.ServicePort),
									},
								},
							},
						}
					}
				}
			}

		}
	}
	certRes, _ := c.certLister.List()
	tlss := tlsConfig{
		Certificates: make([]certificate, 0),
	}
	for _, tlsCert := range certRes {
		cert := tlsCert.(*core.Certificate)
		if cert.Cert == "" || cert.Info.IsCA || len(cert.Info.Domains) == 0 {
			continue
		}
		crt, _ := base64.StdEncoding.DecodeString(cert.Cert)
		key, _ := base64.StdEncoding.DecodeString(cert.Key)
		tlss.Certificates = append(tlss.Certificates, certificate{
			CertFile: string(crt),
			KeyFile:  string(key),
		})
	}
	cfg := traefikConfig{
		HTTP: &httpConfig{
			Routers:  routers,
			Services: svcs,
		},
		TLS: &tlss,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.data = data
	c.mu.Unlock()
}

func getServiceName(name, namespace string, port int) string {
	return fmt.Sprintf("%s_%s_%d", name, namespace, port)
}

func (c *ingress) server(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if c.data == nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Add("content-type", "application/json")
		w.Write(c.data)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
