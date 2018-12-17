package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim"
	"github.com/ordovicia/kubernetes-simulator/log"
	"github.com/ordovicia/kubernetes-simulator/scheduler"
)

type myFilter struct{}

func (f *myFilter) Filter(pod *v1.Pod, node *v1.Node) (ok bool, err error) { return true, nil }

type myScorer struct{}

func (s *myScorer) Score(pod *v1.Pod, nodes []*v1.Node) (scores []scheduler.NodeScore, weight int, err error) {
	return []scheduler.NodeScore{}, 0, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

// configPath is the path of the config file, defaulting to "sample/config"
var configPath string

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config/sample", "config file (exclusing file extension)")
}

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		// Create a new KubeSim
		kubesim, err := kubesim.NewKubeSimFromConfigPath(configPath)
		if err != nil {
			log.G(context.TODO()).WithError(err).Fatalf("Error creating KubeSim: %s", err.Error())
		}

		// Register plugins
		filter := myFilter{}
		kubesim.RegisterFilter(&filter)

		scorer := myScorer{}
		kubesim.RegisterScorer(&scorer)

		// SIGINT calcels submitPods() and kubesim.Run()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			cancel()
		}()

		// Continuously submit pods
		go submitPods(ctx)

		if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
			log.L.Fatal(err)
		}
	},
}

func submitPods(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// pod := ..
			// kubesim.PodQueue() <- pod
			time.Sleep(1 * time.Second)
		}
	}
}
