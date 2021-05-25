package nginx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
)

type ingress struct {
	listenerLister core.ResourceLister
	routeLister    core.ResourceLister
	certLister     core.ResourceLister
	version        string
	data           []byte
	mu             *sync.Mutex
	cfg            *core.IngressConfig
}

type ingressRule struct {
	namespace string
	core.IngressRule
}

func New(listenerLister, routeLister, certLister core.ResourceLister, cfg *core.IngressConfig) core.IngressControllerHandle {
	h := &ingress{
		listenerLister: listenerLister,
		routeLister:    routeLister,
		certLister:     certLister,
		mu:             new(sync.Mutex),
		cfg:            cfg,
	}
	h.handle()
	return h
}

func (c *ingress) UpdateHandle() {
	routeList, rok := c.routeLister.List()
	if !rok {
		return
	}
	listenerList, lok := c.listenerLister.List()
	if !lok {
		return
	}
	certList, cok := c.certLister.List()
	if !cok {
		return
	}

	//
	rules := make(map[string]map[string][]ingressRule)
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
	listen := Listen{
		TCP:  make([]TCPConfig, 0),
		UDP:  make([]UDPConfig, 0),
		HTTP: make([]HTTPConfig, 0),
		TLS:  make(map[string]Cert),
	}

	backends := make([]*Backend, 0)
	for _, v := range listenerList {
		lis := v.(*core.IngressListener)
		if lis.Drive == "" {
			lis.Drive = core.IngressNginxDrive
		}
		if lis.Drive != core.IngressNginxDrive {
			continue
		}

		for host, rules := range rules[lis.Name] {
			tcpFlag := false
			lisAuth := false
			httprs := make([]RouteConfig, 0)
			for _, rule := range rules {
				if rule.TCP != nil {
					name := fmt.Sprintf("%s.%s_%d", rule.TCP.Backend.ServiceName, rule.namespace, rule.TCP.Backend.ServicePort)
					backends = append(backends, &Backend{
						Host: rule.TCP.Backend.ServiceName + "." + rule.namespace,
						Port: rule.TCP.Backend.ServicePort,
					})
					listen.TCP = append(listen.TCP, TCPConfig{
						Port:       lis.Port,
						ServicName: name,
					})
					tcpFlag = true
					break
				}

				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						backends = append(backends, &Backend{
							Host: path.Backend.ServiceName + "." + rule.namespace,
							Port: path.Backend.ServicePort,
						})
						reqAuth := false
						if v, ok := path.Config["auth"]; ok {
							reqAuth = v == "true"
						}
						if reqAuth {
							lisAuth = true
						}
						httprs = append(httprs, RouteConfig{
							EnableAuth: reqAuth,
							Path:       path.Path,
							ServicName: fmt.Sprintf("%s.%s_%d", path.Backend.ServiceName, rule.namespace, path.Backend.ServicePort),
						})
					}

				}
			}
			if !tcpFlag {
				if host == "" {
					host = "localhost"
				}
				listen.HTTP = append(listen.HTTP, HTTPConfig{
					Port:       lis.Port,
					Host:       host,
					Routers:    httprs,
					CertName:   c.getCert(host, certList),
					TLS:        !lis.DisabledTLS,
					EnableAuth: lisAuth,
				})

			}
		}
	}
	for _, v := range certList {
		cert := v.(*core.Certificate)
		crtb, _ := base64.StdEncoding.DecodeString(cert.Cert)
		keyb, _ := base64.StdEncoding.DecodeString(cert.Key)
		if cert.Info.IsCA {
			continue
		}

		listen.TLS[strings.ReplaceAll(cert.Info.Domains[0], "*", "all_")] = Cert{
			Crt: string(crtb),
			Key: string(keyb),
		}
	}
	cfg := IngressConfig{
		Version:  time.Now().String(),
		Backends: backends,
		Listen:   listen,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.data = data
	c.mu.Unlock()
}

func (c *ingress) getCert(host string, certRes []core.Resource) string {
	score := 0
	index := -1
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
	return strings.ReplaceAll(certRes[index].(*core.Certificate).Info.Domains[0], "*", "all_")
}

func (c *ingress) handle() {
	http.HandleFunc("/nginx", func(w http.ResponseWriter, req *http.Request) {
		if c.data == nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Add("content-type", "application/json")
		w.Write(c.data)
	})
}
