package command

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/oars-sigs/oars-cloud/pkg/version"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "oars",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get())
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfg", "", "config file path")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(toolCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
