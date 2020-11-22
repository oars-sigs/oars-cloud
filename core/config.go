package core

//Config 配置
type Config struct {
	Server ServerConfig
	Etcd   EtcdConfig
	Node   NodeConfig
}

//ServerConfig 服务端配置
type ServerConfig struct {
	Port int `envconfig:"SERVER_PORT"  default:"8801"`
	TLS  TLSConfig
}

type TLSConfig struct {
	Enabled  bool   `envconfig:"SERVER_TLS_ENABLED"`
	CAFile   string `envconfig:"SERVER_TLS_CAFILE"`
	CertFile string `envconfig:"SERVER_TLS_CERTFILE"`
	KeyFile  string `envconfig:"SERVER_TLS_KeYFILE"`
}

//EtcdConfig ectd 配置
type EtcdConfig struct {
	TLS           bool     `envconfig:"ETCD_TLS"`
	Prefix        string   `envconfig:"ETCD_PREFIX" default:"oars"`
	Endpoints     []string `envconfig:"ETCD_ENDPOINTS"`
	CertFile      string   `envconfig:"ETCD_CERT_FILE"`
	KeyFile       string   `envconfig:"ETCD_KEY_FILE"`
	TrustedCAFile string   `envconfig:"ETCD_CA_FILE"`
}

type NodeConfig struct {
	Hostname    string   `envconfig:"NODE_HOSTNAME"`
	IP          string   `envconfig:"NODE_IP"`
	Port        int      `envconfig:"NODE_PORT" default:"8802"`
	UpDNS       []string `envconfig:"NODE_UPSTREAN_DNS"`
	MetricsPort int      `envconfig:"NODE_METRUCSPort" default:"8803"`
	WorkDir     string   `envconfig:"NODE_WORKDIR" default:"/opt/oars/woker"`
}
