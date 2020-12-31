package server

import (
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/server/routers"
	"github.com/oars-sigs/oars-cloud/ui"
)

func Start(mgr *core.APIManager) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))
	routers.NewV1(r, mgr)
	uiHandle := MakeGzipHandler(http.FileServer(ui.New()))
	r.GET("/", gin.WrapH(uiHandle))
	r.GET("/app.js", gin.WrapH(uiHandle))
	r.GET("/app.css", gin.WrapH(uiHandle))

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", mgr.Cfg.Server.Port),
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("Listen in :%d \n", mgr.Cfg.Server.Port)
	if mgr.Cfg.Server.TLS.Enabled {
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(mgr.Cfg.Server.TLS.CAFile)
		if err != nil {
			return err
		}
		pool.AppendCertsFromPEM(caCrt)
		// cert, err := tls.LoadX509KeyPair(mgr.Cfg.Server.TLS.CertFile, mgr.Cfg.Server.TLS.KeyFile)
		// if err != nil {
		// 	return err
		// }
		s.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			//Certificates: []tls.Certificate{cert},
			ClientCAs: pool,
		}
		err = s.ListenAndServeTLS(mgr.Cfg.Server.TLS.CertFile, mgr.Cfg.Server.TLS.KeyFile)
		return err
	}

	err := s.ListenAndServe()
	return err
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Use the Writer part of gzipResponseWriter to write the output.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// MakeGzipHandler adds support for gzip compression for given handler
func MakeGzipHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client can accept the gzip encoding.
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}

		// Set the HTTP header indicating encoding.
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler.ServeHTTP(gzw, r)
	})
}
