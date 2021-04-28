package traefik

type traefikConfig struct {
	HTTP *httpConfig `json:"http"`
	TCP  *tcpConfig  `json:"tcp"`
	UDP  *udpConfig  `json:"udp"`
	TLS  *tlsConfig  `json:"tls"`
}

type entryPoint struct {
	Address string `json:"address,omitempty"`
}

type httpConfig struct {
	Routers     map[string]router                 `json:"routers,omitempty"`
	Services    map[string]service                `json:"services,omitempty"`
	Middlewares map[string]map[string]interface{} `json:"middlewares,omitempty"`
}

type tcpConfig struct {
	Routers  map[string]router  `json:"routers,omitempty"`
	Services map[string]service `json:"services,omitempty"`
}

type udpConfig struct {
	Routers  map[string]router  `json:"routers,omitempty"`
	Services map[string]service `json:"services,omitempty"`
}

type router struct {
	EntryPoints []string   `json:"entryPoints,omitempty"`
	Service     string     `json:"service,omitempty"`
	Rule        string     `json:"rule,omitempty"`
	TLS         *routerTLS `json:"tls,omitempty"`
	Middlewares []string   `json:"middlewares,omitempty"`
}

type routerTLS struct {
	Options      string            `json:"options,omitempty"`
	CertResolver string            `json:"certResolver,omitempty"`
	Domains      []routerTLSDomain `json:"domains,omitempty"`
}

type routerTLSDomain struct {
	Main string   `json:"main,omitempty"`
	Sans []string `json:"sans,omitempty"`
}

type service struct {
	LoadBalancer serviceLB `loadBalancer,omitempty`
}

type serviceLB struct {
	Servers []serviceServer `json:"servers,omitempty"`
}

type serviceServer struct {
	URL     string `json:"url,omitempty"`
	Address string `json:"address,omitempty"`
}

type tlsConfig struct {
	Certificates []certificate        `json:"certificates,omitempty"`
	Stores       map[string]tlsStore  `json:"stores,omitempty"`
	Options      map[string]tlsOption `json:"options,omitempty"`
}

type certificate struct {
	CertFile string   `json:"certFile,omitempty"`
	KeyFile  string   `json:"keyFile,omitempty"`
	Stores   []string `json:"stores,omitempty"`
}

type tlsStore struct {
	DefaultCertificate certificate `json:"defaultCertificate,omitempty"`
}

type tlsOption struct {
	ClientAuth tlsClientAuth `json:"clientAuth,omitempty"`
}

type tlsClientAuth struct {
	CAFiles        []string `json:"caFiles,omitempty"`
	clientAuthType string   `json:"clientAuthType,omitempty"`
}

const (
	verifyClientCert = "RequireAndVerifyClientCert"
)
