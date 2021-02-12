module github.com/oars-sigs/oars-cloud

go 1.14

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cloudflare/cfssl v1.5.0
	github.com/containerd/containerd v1.4.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.7.3-0.20190111153827-295413c9d0e1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.3.3 // indirect
	github.com/envoyproxy/go-control-plane v0.9.7
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-acme/lego v2.7.2+incompatible
	github.com/go-acme/lego/v4 v4.2.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/miekg/dns v1.1.31
	github.com/moby/ipvs v1.0.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/shirou/gopsutil/v3 v3.20.10
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/vishvananda/netlink v1.1.0
	go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb
	google.golang.org/grpc v1.27.1
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	software.sslmate.com/src/go-pkcs12 v0.0.0-20201103104416-57fc603b7f52
)

replace (
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // ae9734ed278b is the SHA for git tag v3.4.13
)
