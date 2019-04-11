// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"

	kubesim "github.com/pfnet-research/k8s-cluster-simulator/pkg"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/scheduler"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

// configPath is the path of the config file, defaulting to "config".
var configPath string

func init() {
	rootCmd.PersistentFlags().StringVar(
		&configPath, "config", "config", "config file (excluding file extension)")
}

var rootCmd = &cobra.Command{
	Use:   "k8s-cluster-simulator",
	Short: "k8s-cluster-simulator provides a virtual kubernetes cluster interface for evaluating your scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		ctx := newInterruptableContext()

		// 1. Create a KubeSim with a pod queue and a scheduler.
		queue := queue.NewPriorityQueue()
		sched := buildScheduler() // see below
		kubesim := kubesim.NewKubeSimFromConfigPathOrDie(configPath, queue, sched)

		// 2. Register one or more pod submitters to KubeSim.
		numOfSubmittingPods := 8
		kubesim.AddSubmitter("MySubmitter", newMySubmitter(numOfSubmittingPods))

		// 3. Run the main loop of KubeSim.
		//    In each execution of the loop, KubeSim
		//      1) stores pods submitted from the registered submitters to its queue,
		//      2) invokes scheduler with pending pods and cluster state,
		//      3) emits cluster metrics to designated location(s) if enabled
		//      4) progresses the simulated clock
		if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
			log.L.Fatal(err)
		}
	},
}

func buildScheduler() scheduler.Scheduler {
	// 1. Create a generic scheduler that mimics a kube-scheduler.
	sched := scheduler.NewGenericScheduler( /* preemption enabled */ true)

	// 2. Register extender(s)
	sched.AddExtender(
		scheduler.Extender{
			Name:             "MyExtender",
			Filter:           filterExtender,
			Prioritize:       prioritizeExtender,
			Weight:           1,
			NodeCacheCapable: true,
		},
	)

	// 2. Register plugin(s)
	// Predicate
	sched.AddPredicate("GeneralPredicates", predicates.GeneralPredicates)
	// Prioritizer
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "BalancedResourceAllocation",
		Map:    priorities.BalancedResourceAllocationMap,
		Reduce: nil,
		Weight: 1,
	})
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "LeastRequested",
		Map:    priorities.LeastRequestedPriorityMap,
		Reduce: nil,
		Weight: 1,
	})

	return &sched
}

func newInterruptableContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// SIGINT (Ctrl-C) and SIGTERM cancel kubesim.Run().
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	return ctx
}

// for test
func lifo(pod0, pod1 *v1.Pod) bool { // nolint
	return pod1.CreationTimestamp.Before(&pod0.CreationTimestamp)
}
