package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type Func func(args interface{}) *core.APIReply

type Server struct {
	path     string
	addr     string
	funcs    map[string]Func
	caFile   string
	certFile string
	keyFile  string
}

func NewServer(addr, path, caFile, certFile, keyFile string) *Server {
	return &Server{
		path:     path,
		funcs:    make(map[string]Func),
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Register a method via name
func (s *Server) Register(name string, f Func) {
	if _, ok := s.funcs[name]; ok {
		return
	}
	s.funcs[name] = f
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.write(w, e.InvalidParameterError(err))
		return
	}
	var input core.APIInput
	err = json.Unmarshal(data, &input)
	if err != nil {
		s.write(w, e.InvalidParameterError(err))
		return
	}
	fn, ok := s.funcs[input.Method]
	if !ok {
		s.write(w, e.MethodNotFoundError(err))
		return
	}
	reply := fn(input.Args)
	if err != nil {
		s.write(w, e.InvalidParameterError(err))
		return
	}
	s.write(w, reply)
}

func (s *Server) write(w http.ResponseWriter, reply *core.APIReply) {
	data, _ := json.Marshal(reply)
	w.Write(data)
}

func UnmarshalArgs(in, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, out)
	return err
}

func (s *Server) Listen() error {
	pool := x509.NewCertPool()
	crt, err := ioutil.ReadFile(s.caFile)
	if err != nil {
		return err
	}
	pool.AppendCertsFromPEM(crt)
	http.Handle(s.path, s)
	ser := &http.Server{
		Addr: s.addr,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	return ser.ListenAndServeTLS(s.certFile, s.keyFile)
}
