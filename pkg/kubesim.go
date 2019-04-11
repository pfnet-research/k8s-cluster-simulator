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

package kubesim

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/log"
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/config"
	l "github.com/pfnet-research/k8s-cluster-simulator/pkg/log"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/scheduler"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// KubeSim represents a simulated kubernetes cluster.
type KubeSim struct {
	tick  time.Duration
	clock clock.Clock

	nodes       map[string]*node.Node
	pendingPods queue.PodQueue
	boundPods   map[string]*pod.Pod

	submitters map[string]submitter.Submitter
	scheduler  scheduler.Scheduler

	metricsWriters []metrics.Writer
	metricsTick    time.Duration
}

// NewKubeSim creates a new KubeSim with the given config, queue, and scheduler.
// Returns error if the configuration failed.
func NewKubeSim(
	conf *config.Config, queue queue.PodQueue, sched scheduler.Scheduler,
) (*KubeSim, error) {

	log.G(context.TODO()).Debugf("Config: %+v", *conf)

	if err := configLog(conf.LogLevel); err != nil {
		return nil, errors.Errorf("Error configuring logging: %s", err.Error())
	}

	clk, err := buildClock(conf.StartClock)
	if err != nil {
		return nil, err
	}

	nodes, err := buildCluster(conf)
	if err != nil {
		return nil, err
	}

	metricsTick := conf.Tick
	if conf.MetricsTick != 0 {
		metricsTick = conf.MetricsTick
	}

	metricsWriters, err := buildMetricsWriters(conf)
	if err != nil {
		return nil, err
	}

	return &KubeSim{
		tick:  time.Duration(conf.Tick) * time.Second,
		clock: clk,

		nodes:       nodes,
		pendingPods: queue,
		boundPods:   map[string]*pod.Pod{},

		submitters: map[string]submitter.Submitter{},
		scheduler:  sched,

		metricsTick:    time.Duration(metricsTick) * time.Second,
		metricsWriters: metricsWriters,
	}, nil
}

// NewKubeSimFromConfigPath creates a new KubeSim with config from confPath (excluding file
// extension), queue, and scheduler.
// Returns error if the configuration failed.
func NewKubeSimFromConfigPath(
	confPath string, queue queue.PodQueue, sched scheduler.Scheduler,
) (*KubeSim, error) {

	conf, err := readConfig(confPath)
	if err != nil {
		return nil, errors.Errorf("Error reading config: %s", err.Error())
	}

	return NewKubeSim(conf, queue, sched)
}

// NewKubeSimFromConfigPathOrDie creates a new KubeSim with config from confPath (excluding file
// extension), queue, and scheduler.
// If an error occurs during the initialization, it panics and stops the execution.
func NewKubeSimFromConfigPathOrDie(
	confPath string, queue queue.PodQueue, sched scheduler.Scheduler,
) *KubeSim {

	kubesim, err := NewKubeSimFromConfigPath(confPath, queue, sched)
	if err != nil {
		log.L.Fatal(err)
	}

	return kubesim
}

// AddSubmitter adds the new submitter to this KubeSim.
func (k *KubeSim) AddSubmitter(name string, submitter submitter.Submitter) {
	k.submitters[name] = submitter
}

// Run executes the main loop, which invokes submitters and the scheduler, and binds pods to the
// selected nodes.
// This method blocks until ctx is done or this KubeSim finishes processing all pods.
func (k *KubeSim) Run(ctx context.Context) error {
	preMetricsClock := k.clock
	met, err := metrics.BuildMetrics(k.clock, k.nodes, k.pendingPods)
	if err != nil {
		return err
	}

	submitterAddedEver := len(k.submitters) > 0

	for {
		if k.toTerminate(submitterAddedEver) {
			log.L.Debug("Terminate KubeSim")
			break
		}
		submitterAddedEver = submitterAddedEver || len(k.submitters) > 0

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.L.Debugf("Clock %s", k.clock.ToRFC3339())

			if k.submit(met) != nil {
				return err
			}

			if k.schedule() != nil {
				return err
			}

			// Rebuild metrics every tick for submitters to use.
			met, err = metrics.BuildMetrics(k.clock, k.nodes, k.pendingPods)
			if err != nil {
				return err
			}

			if k.clock.Sub(preMetricsClock) > k.metricsTick {
				preMetricsClock = k.clock
				if err = k.writeMetrics(&met); err != nil {
					return err
				}

				k.gcTerminatedPodsInNodes()
			}

			k.clock = k.clock.Add(k.tick)
		}
	}

	return nil
}

// List implements "k8s.io/pkg/scheduler/algorithm".NodeLister interface.
// Never returns an error.
func (k *KubeSim) List() ([]*v1.Node, error) {
	nodes := make([]*v1.Node, 0, len(k.nodes))
	for _, node := range k.nodes {
		nodes = append(nodes, node.ToV1())
	}
	return nodes, nil
}

// readConfig reads and parses a config from the path (excluding file extension).
func readConfig(path string) (*config.Config, error) {
	viper.SetConfigName(path)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	log.G(context.TODO()).Debugf("Config file %s", viper.ConfigFileUsed())

	var conf = config.Config{
		LogLevel: "info",
		Tick:     10,
	}

	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func configLog(logLevel string) error {
	level, err := log.ParseLevel(logLevel) // if logLevel == "" then level <- info
	if err != nil {
		return strongerrors.InvalidArgument(
			errors.Errorf("Log level %q not supported: %s", level, err.Error()))
	}
	logrus.SetLevel(level)

	logger := log.L
	log.L = logger

	return nil
}

func buildClock(startClock string) (clock.Clock, error) {
	clk := clock.NewClock(time.Now())

	if startClock != "" {
		c, err := time.Parse(time.RFC3339, startClock)
		if err != nil {
			return clk, err
		}
		clk = clock.NewClock(c)
	}

	return clk, nil
}

func buildCluster(conf *config.Config) (map[string]*node.Node, error) {
	nodes := map[string]*node.Node{}
	for _, nodeConf := range conf.Cluster {
		nodeV1, err := config.BuildNode(nodeConf, conf.StartClock)
		if err != nil {
			return nil, err
		}

		nodeSim := node.NewNode(nodeV1)
		nodes[nodeV1.Name] = &nodeSim

		log.L.Debugf("Node %s created: %v", nodeV1.Name, nodeV1)
	}

	return nodes, nil
}

func buildMetricsWriters(conf *config.Config) ([]metrics.Writer, error) {
	writers := []metrics.Writer{}

	fileWriters, err := config.BuildMetricsLogger(conf.MetricsLogger)
	if err != nil {
		return []metrics.Writer{}, err
	}

	for _, writer := range fileWriters {
		log.L.Infof("Metrics and log written to %s", writer.FileName())
		writers = append(writers, writer)
	}

	return writers, nil
}

// toTerminate determines whether the main loop of this KubeSim can be terminated,
// because all submitters are terminated, no pods are running on the cluster, and there are no
// pending pods in the queue.
func (k *KubeSim) toTerminate(submitterAddedEver bool) bool {
	if _, err := k.pendingPods.Front(); err == queue.ErrEmptyQueue { // queue is empty
		for _, node := range k.nodes { // cluster is empty
			if node.PodsNum(k.clock) > 0 {
				return false
			}
		}

		if submitterAddedEver && len(k.submitters) == 0 { // all submitters are terminated
			return true
		}
	}

	return false
}

func (k *KubeSim) submit(metrics metrics.Metrics) error {
	for name, subm := range k.submitters {
		events, err := subm.Submit(k.clock, k, metrics)
		if err != nil {
			return err
		}

		for _, e := range events {
			if submitted, ok := e.(*submitter.SubmitEvent); ok {
				pod := submitted.Pod
				pod.UID = types.UID(pod.Name) // FIXME
				pod.CreationTimestamp = k.clock.ToMetaV1()
				pod.Status.Phase = v1.PodPending

				log.L.Tracef("Submitter %s: Submit %v", name, pod)

				if l.IsDebugEnabled() {
					key, err := util.PodKey(pod)
					if err != nil {
						return err
					}
					log.L.Debugf("Submitter %s: Submit %s", name, key)
				}

				err := k.pendingPods.Push(pod)
				if err != nil {
					return err
				}
			} else if del, ok := e.(*submitter.DeleteEvent); ok {
				log.L.Debugf("Submitter %s: Delete %s",
					name, util.PodKeyFromNames(del.PodNamespace, del.PodName))

				if delFromQ := k.pendingPods.Delete(del.PodNamespace, del.PodName); !delFromQ {
					k.deletePodFromNode(del.PodNamespace, del.PodName)
				}
			} else if up, ok := e.(*submitter.UpdateEvent); ok {
				log.L.Tracef("Submitter %s: Update %s to %v",
					name, util.PodKeyFromNames(up.PodNamespace, up.PodName), up.NewPod)
				log.L.Debugf("Submitter %s: Update %s",
					name, util.PodKeyFromNames(up.PodNamespace, up.PodName))

				if err := k.pendingPods.Update(up.PodNamespace, up.PodName, up.NewPod); err != nil {
					if e, ok := err.(*queue.ErrNoMatchingPod); ok {
						log.L.Warnf("Error updating pod: %s", e.Error())
					} else {
						return err
					}
				}
			} else if _, ok := e.(*submitter.TerminateSubmitterEvent); ok {
				log.L.Debugf("Submitter %s: Terminate", name)
				delete(k.submitters, name)
			} else {
				log.L.Panic("Unknown submitter event")
			}
		}
	}

	return nil
}

func (k *KubeSim) schedule() error {
	// Build up-to-date NodeInfo.
	nodeInfoMap := make(map[string]*nodeinfo.NodeInfo, len(k.nodes))
	for name, node := range k.nodes {
		info, err := node.ToNodeInfo(k.clock)
		if err != nil {
			return err
		}
		nodeInfoMap[name] = info
	}

	// The scheduler makes scheduling decision.
	events, err := k.scheduler.Schedule(k.clock, k.pendingPods, k, nodeInfoMap)
	if err != nil {
		return err
	}

	// Do the actual scheduling process for each event.
	for _, e := range events {
		if bind, ok := e.(*scheduler.BindEvent); ok {
			nodeName := bind.ScheduleResult.SuggestedHost
			node, ok := k.nodes[nodeName]
			if !ok {
				return fmt.Errorf("No node named %q", nodeName)
			}
			bind.Pod.Spec.NodeName = nodeName

			pod, err := node.BindPod(k.clock, bind.Pod)
			if err != nil {
				return err
			}

			key, err := util.PodKey(bind.Pod)
			if err != nil {
				return err
			}
			k.boundPods[key] = pod
		} else if del, ok := e.(*scheduler.DeleteEvent); ok {
			k.deletePodFromNode(del.PodNamespace, del.PodName)
		} else {
			log.L.Panic("Unknown scheduler event")
		}
	}

	return nil
}

func (k *KubeSim) writeMetrics(met *metrics.Metrics) error {
	for _, writer := range k.metricsWriters {
		if err := writer.Write(met); err != nil {
			return err
		}
	}

	return nil
}

func (k *KubeSim) gcTerminatedPodsInNodes() {
	for _, node := range k.nodes {
		node.GCTerminatedPods(k.clock)
	}
}

func (k *KubeSim) deletePodFromNode(podNamespace, podName string) {
	key := util.PodKeyFromNames(podNamespace, podName)
	k.boundPods[key].Delete(k.clock)

	nodeName := k.boundPods[key].ToV1().Spec.NodeName
	deletedFromNode := k.nodes[nodeName].DeletePod(k.clock, podNamespace, podName) // nolint

	if !deletedFromNode { // nolint
		//
	}
}
