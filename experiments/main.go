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
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

const (
	BEST_FIT = "bestfit"
	OVER_SUB = "oversub"
	PROPOSED = "proposed"
	GENERTIC = "generic"
)

// configPath is the path of the config file, defaulting to "config".
var (
	configPath    string
	isGenWorkload = true
	// isGenWorkload    = false
	workloadPath     string
	targetNum        = 64
	totalPodsNum     = uint64(640 * 5)
	submittedPodsNum = uint64(0)
	podMap           = make(map[string][]string)
	// schedulerName    = "bestfit"
	schedulerName = "default"
	// schedulerName       = "proposed"
	globalOverSubFactor = 4.0

	meanSec    = 10.0
	meanCpu    = 4.0
	phasNum    = 1
	requestCpu = 8.0
)

func init() {
	log.L.Infof("Running KubeSim @ %s", time.Now().Format(time.RFC850))
	rootCmd.PersistentFlags().StringVar(
		&configPath, "config", "./config/c2node", "config file (excluding file extension)")
	rootCmd.PersistentFlags().StringVar(
		&workloadPath, "workload", "./config/workload", "config file (excluding file extension)")
	rootCmd.PersistentFlags().BoolVar(
		&isGenWorkload, "isgen", false, "generating workload")
	rootCmd.PersistentFlags().StringVar(
		&schedulerName, "scheduler", "default", "generating workload")
	rootCmd.PersistentFlags().Float64Var(
		&globalOverSubFactor, "oversub", 1.0, "generating workload")
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
		nodes, _ := kubesim.List()
		for _, node := range nodes {
			predicates.NodesOverSubFactors[node.Name] = globalOverSubFactor
		}
		// 2. Prepare the set of podsubmit time: set<timestamp>

		// 3. Register one or more pod submitters to KubeSim.
		kubesim.AddSubmitter("MySubmitter", newMySubmitter(totalPodsNum))

		// 4. Run the main loop of KubeSim.
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
	if !isGenWorkload {
		files, _ := ioutil.ReadDir(workloadPath)
		totalPodsNum = uint64(len(files))
		for _, f := range files {
			fileName := string(f.Name())
			arr := strings.Split(fileName, "@")
			clockStr := arr[0]
			podName := arr[1]
			if _, ok := podMap[clockStr]; ok {
				podMap[clockStr] = append(podMap[clockStr], podName)
			} else {
				strArr := []string{podName}
				podMap[clockStr] = strArr
			}
		}
	} else {
		os.RemoveAll(workloadPath)
		os.MkdirAll(workloadPath, 0755)
	}
	log.L.Infof("scheduler input %s", schedulerName)
	log.L.Infof("Submitting %d pods", totalPodsNum)
	log.L.Infof("workload: %s", workloadPath)
	log.L.Infof("isGenWorkload: %v", isGenWorkload)
	log.L.Infof("cluster: %s", configPath)
	log.L.Infof("oversub: %f", globalOverSubFactor)
	switch schedName := strings.ToLower(schedulerName); schedName {

	case PROPOSED:
		log.L.Infof("Scheduler: %s", PROPOSED)
		globalOverSubFactor = 2.0
		sched := scheduler.NewProposedScheduler(false)
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
		sched.AddPredicate("PodFitsResourcesOverSub", predicates.PodFitsResourcesOverSub)
		// Prioritizer
		sched.AddPrioritizer(priorities.PriorityConfig{
			Name:   "MostRequested",
			Map:    priorities.MostRequestedPriorityMap,
			Reduce: nil,
			Weight: 1,
		})

		return &sched
	case OVER_SUB:
		log.L.Infof("Scheduler: %s", OVER_SUB)
		sched := scheduler.NewGenericScheduler(false)
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
		sched.AddPredicate("PodFitsResourcesOverSub", predicates.PodFitsResourcesOverSub)
		// Prioritizer
		sched.AddPrioritizer(priorities.PriorityConfig{
			Name:   "MostRequested",
			Map:    priorities.MostRequestedPriorityMap,
			Reduce: nil,
			Weight: 1,
		})

		return &sched
	case BEST_FIT:
		log.L.Infof("Scheduler: %s", BEST_FIT)
		globalOverSubFactor = 1.0
		sched := scheduler.NewGenericScheduler(false)
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
		sched.AddPredicate("PodFitsResources", predicates.PodFitsResources)
		// Prioritizer
		sched.AddPrioritizer(priorities.PriorityConfig{
			Name:   "MostRequested",
			Map:    priorities.MostRequestedPriorityMap,
			Reduce: nil,
			Weight: 1,
		})

		return &sched
	default:
		log.L.Infof("Scheduler: DEFAULT")
		// 1. Create a generic scheduler that mimics a kube-scheduler.
		sched := scheduler.NewGenericScheduler( /* preemption disabled */ false)
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
