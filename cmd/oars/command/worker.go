package command

import (
	"fmt"
	"os"
	"time"

	"github.com/oars-sigs/oars-cloud/pkg/config"
	"github.com/oars-sigs/oars-cloud/pkg/etcd"
	"github.com/oars-sigs/oars-cloud/pkg/version"
	"github.com/oars-sigs/oars-cloud/pkg/worker"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use: "worker",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get())
		cfg, err := config.Load(cfgFile)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
		store, err := etcd.New(&cfg.Etcd, 5*time.Second)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
		err = worker.Start(store, cfg.Node)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	},
}
