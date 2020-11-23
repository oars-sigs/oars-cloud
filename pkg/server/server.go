package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/server/routers"
	"github.com/oars-sigs/oars-cloud/ui"
)

func Start(mgr *core.APIManager) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	routers.NewV1(r, mgr)
	r.GET("/", gin.WrapH(http.FileServer(ui.New())))
	r.GET("/app.js", gin.WrapH(http.FileServer(ui.New())))
	r.GET("/app.css", gin.WrapH(http.FileServer(ui.New())))

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
