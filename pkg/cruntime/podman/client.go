package podman

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type client struct {
	baseUrl    string
	httpClient *http.Client
}

type ClientConfig struct {
	SocketPath  string
	HttpTimeout time.Duration
}

func DefaultClientConfig() ClientConfig {
	cfg := ClientConfig{
		HttpTimeout: 60 * time.Second,
	}
	uid := os.Getuid()
	// are we root?
	if uid == 0 {
		cfg.SocketPath = "unix:/run/podman/podman.sock"
	} else {
		// not? then let's try the default per-user socket location
		cfg.SocketPath = fmt.Sprintf("unix:/run/user/%d/podman/podman.sock", uid)
	}
	return cfg
}

func NewClient(config ClientConfig) *client {
	c := &client{}

	baseUrl := config.SocketPath
	c.httpClient = &http.Client{
		Timeout: config.HttpTimeout,
	}
	if strings.HasPrefix(baseUrl, "unix:") {
		c.baseUrl = "http://u"
		path := strings.TrimPrefix(baseUrl, "unix:")
		c.httpClient.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		}
	} else {
		c.baseUrl = baseUrl
	}

	return c
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	res, err := c.httpClient.Do(req)
	return res, err
}

func (c *client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}

func (c *client) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
