package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/ordovicia/kubernetes-simulator/kubesim"
	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/scheduler"
	"github.com/ordovicia/kubernetes-simulator/log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

// configPath is the path of the config file, defaulting to "examples/config_sample".
var configPath string

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "examples/config_sample", "config file (exclusing file extension)")
}

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for evaluating your scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		// Create a new KubeSim
		kubesim, err := kubesim.NewKubeSimFromConfigPath(configPath)
		if err != nil {
			log.G(context.TODO()).WithError(err).Fatalf("Error creating KubeSim: %s", err.Error())
		}

		// Register a submitter
		submitter := mySubmitter{}
		kubesim.RegisterSubmitter(&submitter)

		// Add an extender
		kubesim.Scheduler().AddExtender(
			scheduler.Extender{
				Name:             "MyExtender",
				Filter:           filter,
				Prioritize:       prioritize,
				Weight:           1,
				NodeCacheCapable: true,
			},
		)

		// SIGINT cancels the sumbitter and kubesim.Run().
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

// Submitter
type mySubmitter struct {
	startClock clock.Clock
	n          uint64
}

func (s *mySubmitter) Submit(clock clock.Clock, nodes []*v1.Node) (pods []*v1.Pod, err error) {
	if s.n == 0 {
		s.startClock = clock
	}

	elapsed := clock.Sub(s.startClock).Seconds()
	if uint64(elapsed)/5 >= s.n {
		pod := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              fmt.Sprintf("pod-%d", s.n),
				Namespace:         "default",
				CreationTimestamp: clock.ToMetaV1(),
				Annotations: map[string]string{
					"simSpec": `
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10
  resourceUsage:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
`,
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Name:  "container",
						Image: "container",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								"cpu":            resource.MustParse("3"),
								"memory":         resource.MustParse("5Gi"),
								"nvidia.com/gpu": resource.MustParse("1"),
							},
							Limits: v1.ResourceList{
								"cpu":            resource.MustParse("4"),
								"memory":         resource.MustParse("6Gi"),
								"nvidia.com/gpu": resource.MustParse("1"),
							},
						},
					},
				},
			},
		}

		s.n++
		return []*v1.Pod{&pod}, nil
	}

	return []*v1.Pod{}, nil
}

func filter(args api.ExtenderArgs) api.ExtenderFilterResult {
	return api.ExtenderFilterResult{
		Nodes:       &v1.NodeList{},
		NodeNames:   args.NodeNames,
		FailedNodes: api.FailedNodesMap{},
		Error:       "",
	}
}

func prioritize(args api.ExtenderArgs) api.HostPriorityList {
	priorities := api.HostPriorityList{}
	for _, name := range *args.NodeNames {
		priorities = append(priorities, api.HostPriority{Host: name, Score: 1})
	}

	return priorities
}
