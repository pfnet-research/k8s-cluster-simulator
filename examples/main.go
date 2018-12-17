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

// Filter
type myFilter struct{}

func (f *myFilter) Filter(pod *v1.Pod, node *v1.Node) (ok bool, err error) {
	resourceReq := v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		resourceReq = resourceSum(resourceReq, container.Resources.Requests)
	}
	return resourceGE(node.Status.Allocatable, resourceReq), nil
}

func resourceSum(r1, r2 v1.ResourceList) v1.ResourceList {
	sum := r1
	for r2Key := range r2 {
		if r2Val, ok := sum[r2Key]; ok {
			r1Val := sum[r2Key]
			r1Val.Add(r2Val)
			sum[r2Key] = r1Val
		} else {
			sum[r2Key] = r2Val
		}
	}
	return sum
}

func resourceGE(r1, r2 v1.ResourceList) bool {
	for r2Key, r2Val := range r2 {
		if r1Val, ok := r1[r2Key]; !ok {
			return false
		} else if r1Val.Cmp(r2Val) <= 0 {
			return false
		}
	}
	return true
}

// Scorer
type myScorer struct{}

func (s *myScorer) Score(pod *v1.Pod, nodes []*v1.Node) (scores []scheduler.NodeScore, weight int, err error) {
	scores = []scheduler.NodeScore{}
	for _, node := range nodes {
		scores = append(scores, scheduler.NodeScore{Node: node.Name, Score: 1})
	}
	return scores, 1, nil
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
