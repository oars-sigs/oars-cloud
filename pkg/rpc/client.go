package rpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"

	"github.com/gorilla/websocket"
)

type Client struct {
	client   *http.Client
	wsDailer *websocket.Dialer
	path     string
	caFile   string
	certFile string
	keyFile  string
}

func NewClient(path, caFile, certFile, keyFile string) (*Client, error) {
	c := &Client{
		path:     path,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
	}
	err := c.newClient()
	return c, err
}

func (c *Client) newClient() error {
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(c.caFile)
	if err != nil {
		return err
	}
	pool.AppendCertsFromPEM(caCrt)
	clientCrt, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{clientCrt},
	}
	tr := &http.Transport{
		TLSClientConfig:   tlsConfig,
		DisableKeepAlives: true,
	}
	httpc := &http.Client{
		Transport: tr,
	}
	c.client = httpc
	c.wsDailer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  tlsConfig,
	}
	return nil
}

func (c *Client) WSDailer() *websocket.Dialer {
	return c.wsDailer
}

func (c *Client) Call(addr, method string, args interface{}) *core.APIReply {
	in := core.APIInput{
		Method: method,
		Args:   args,
	}
	body, err := json.Marshal(in)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	resp, err := c.client.Post("https://"+addr+c.path, "application/json ", bytes.NewBuffer(body))
	if err != nil {
		return e.InternalError(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return e.InternalError(err)
	}
	var reply core.APIReply
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return e.InternalError(err)
	}

	return &reply
}
