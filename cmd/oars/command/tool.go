package command

import (
	"io/ioutil"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/utils/rsa"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use: "tool",
	Run: func(cmd *cobra.Command, args []string) {
		caCert, err := ioutil.ReadFile(args[0])
		if err != nil {
			panic(err)
		}
		caKey, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		rootCrt, err := rsa.ParseCrt(caCert)
		if err != nil {
			panic(err)
		}
		key, err := rsa.ParseKey(caKey)
		if err != nil {
			panic(err)
		}
		serverInfo := &core.CertInformation{
			CommonName:  "Worker",
			IPAddresses: []string{args[2]},
			Domains:     []string{"localhost"},
		}
		serverCrt, serverKey, err := rsa.CreateCRT(rootCrt, key, serverInfo)
		if err != nil {
			panic(err)
		}
		ioutil.WriteFile(args[3]+".crt", serverCrt, 0664)
		ioutil.WriteFile(args[3]+".key", serverKey, 0664)
	},
}
