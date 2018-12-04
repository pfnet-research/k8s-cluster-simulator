package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ordovicia/kubernetes-simulator/kubesim"
	"github.com/ordovicia/kubernetes-simulator/log"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Long:  "FIXME: kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		_ = ctx

		kubesim, err := kubesim.NewKubeSim(configPath, nil)
		if err != nil {
			log.G(context.TODO()).WithError(err).Fatal("Error creating KubeSim: %s", err.Error())
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			cancel()
		}()

		if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
			log.L.Fatal(err)
		}
	},
}

// Execute executes the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (exclusing file extension)")
}
