package controller

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/acme"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"

	log "github.com/sirupsen/logrus"
)

type certController struct {
	store      core.KVStore
	certStore  core.ResourceStore
	certLister core.ResourceLister
}

func newCert(kv core.KVStore) (*certController, error) {
	certLister, err := resStore.NewLister(kv, &core.Certificate{}, &core.ResourceEventHandle{})
	if err != nil {
		return nil, err
	}
	return &certController{
		certStore:  resStore.NewStore(kv, new(core.Certificate)),
		certLister: certLister,
	}, nil
}

func (c *certController) run() {
	t := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-t.C:
			err := c.update()
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (c *certController) update() error {
	resources, _ := c.certLister.List()
	for _, resource := range resources {
		cert := resource.(*core.Certificate)
		if cert.Acme == nil || cert.Info == nil {
			continue
		}
		if cert.Acme.Account.Registration == nil {
			err := acme.NewAccount(cert.Acme.Account)
			if err != nil {
				log.Error(err)
				continue
			}
			_, err = c.certStore.Put(context.Background(), cert, &core.PutOptions{})
			if err != nil {
				log.Error(err)
				continue
			}
		}
		if cert.Cert == "" {
			log.Info("create cert", cert.Info.Domains[0])
			cli, err := acme.New(cert)
			if err != nil {
				log.Error(err)
				continue
			}
			cert, err = cli.Create()
			if err != nil {
				log.Error(err)
				continue
			}
		}
		if cert.Cert != "" {
			if cert.Info.NotAfter.After(time.Now().AddDate(0, 1, 0)) {
				continue
			}
			log.Info("renew cert", cert.Info.Domains[0])
			cli, err := acme.New(cert)
			if err != nil {
				log.Error(err)
				continue
			}
			cert, err = cli.Renew()
			if err != nil {
				log.Error(err)
				continue
			}
		}
		_, err := c.certStore.Put(context.Background(), cert, &core.PutOptions{})
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}
