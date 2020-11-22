package agent

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/sirupsen/logrus"
)

func (d *daemon) dnsServer() {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", d.dnsHandle)
	server := &dns.Server{
		Addr:         ":53",
		Net:          "udp",
		Handler:      handler,
		UDPSize:      65535,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logrus.Infof("Startlistener on %s", ":53")
	err := server.ListenAndServe()
	if err != nil {
		logrus.Error(err)
	}
}

func (d *daemon) dnsHandle(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	domain := q.Name
	dom := strings.TrimSuffix(domain, "DHCP\\ HOST.")
	dom = strings.TrimSuffix(dom, ".")
	if q.Qtype == dns.TypeA {
		addrs := make([]string, 0)
		d.endpointCache.Range(func(k, v interface{}) bool {
			if endpoint, ok := v.(*core.Endpoint); ok {
				if endpoint.Service+"."+endpoint.Namespace == dom {
					addrs = append(addrs, endpoint.HostIP)
				}
			}
			return true
		})
		if len(addrs) == 0 {
			cli := &dns.Client{
				Net:          "udp",
				UDPSize:      65535,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			}
			for _, dn := range d.node.UpDNS {
				m, _, err := cli.Exchange(req, dn+":53")
				if err != nil {
					logrus.Error(err)
					continue
				}
				w.WriteMsg(m)
				return
			}
			dns.HandleFailed(w, req)
			return
		}
		m := new(dns.Msg)
		m.SetReply(req)
		m.Authoritative = true
		m.Answer = make([]dns.RR, 0)
		for _, addr := range addrs {
			rr := new(dns.A)
			rr.Hdr = dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 86400}
			rr.A = net.ParseIP(addr)
			m.Answer = append(m.Answer, rr)
		}
		w.WriteMsg(m)
	}
}
