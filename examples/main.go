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
	sched "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/ordovicia/kubernetes-simulator/kubesim"
	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/log"
)

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

		// Register submitter
		submitter := mySubmitter{}
		kubesim.RegisterSubmitter(&submitter)

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

// Filter
type myFilter struct{}

func (f *myFilter) Filter(pod *v1.Pod, node *v1.Node) (ok bool, err error) {
	return true, nil
}

// Scorer
type myScorer struct{}

func (s *myScorer) Score(pod *v1.Pod, nodes []*v1.Node) (scores sched.HostPriorityList, weight int, err error) {
	for _, node := range nodes {
		scores = append(scores, sched.HostPriority{Host: node.Name, Score: 1})
	}

	return scores, 1, nil
}
