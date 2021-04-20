package nginx

type IngressConfig struct {
	Version  string
	Listen   Listen
	Backends []*Backend
}

type Listen struct {
	HTTP []HTTPConfig
	TCP  []TCPConfig
	UDP  []UDPConfig
	TLS  map[string]Cert
}

type Cert struct {
	Crt string
	Key string
}

type HTTPConfig struct {
	Port     int
	Host     string
	CertName string
	Routers  []RouteConfig
	TLS      bool
}

type RouteConfig struct {
	Path       string
	ServicName string
}

type TCPConfig struct {
	Port       int
	ServicName string
}

type UDPConfig struct {
	Port       int
	ServicName string
}

type Backend struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
	Port int    `json:"port"`
}
