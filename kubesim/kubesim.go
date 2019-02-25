package kubesim

import (
	"context"
	"time"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/api"
	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/config"
	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	"github.com/ordovicia/kubernetes-simulator/kubesim/scheduler"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// KubeSim represents a kubernetes cluster simulator.
type KubeSim struct {
	nodes    map[string]*node.Node
	podQueue queue.PodQueue

	tick  int
	clock clock.Clock

	submitters []api.Submitter
	scheduler  scheduler.Scheduler
}

// NewKubeSim creates a new KubeSim with the given config, queue, and scheduler.
func NewKubeSim(conf *config.Config, queue queue.PodQueue, sched scheduler.Scheduler) (*KubeSim, error) {
	log.G(context.TODO()).Debugf("Config: %+v", *conf)

	if err := configLog(conf.LogLevel); err != nil {
		return nil, errors.Errorf("error configuring: %s", err.Error())
	}

	clk := time.Now()
	if conf.StartClock != "" {
		var err error
		clk, err = time.Parse(time.RFC3339, conf.StartClock)
		if err != nil {
			return nil, err
		}
	}

	nodes := map[string]*node.Node{}
	for _, nodeConf := range conf.Cluster.Nodes {
		log.L.Debugf("NodeConfig: %+v", nodeConf)

		nodeV1, err := config.BuildNode(nodeConf, conf.StartClock)
		if err != nil {
			return nil, errors.Errorf("error building node config: %s", err.Error())
		}

		n := node.NewNode(nodeV1)
		nodes[nodeV1.Name] = &n

		log.L.Debugf("Node %q created", nodeV1.Name)
	}

	kubesim := KubeSim{
		nodes:     nodes,
		podQueue:  queue,
		tick:      conf.Tick,
		clock:     clock.NewClock(clk),
		scheduler: sched,
	}

	return &kubesim, nil
}

// NewKubeSimFromConfigPath creates a new KubeSim with config from confPath (excluding file path),
// queue, and scheduler.
func NewKubeSimFromConfigPath(confPath string, queue queue.PodQueue, sched scheduler.Scheduler) (*KubeSim, error) {
	conf, err := readConfig(confPath)
	if err != nil {
		return nil, errors.Errorf("error reading config: %s", err.Error())
	}

	return NewKubeSim(conf, queue, sched)
}

// AddSubmitter adds a new submitter plugin to this KubeSim.
func (k *KubeSim) AddSubmitter(submitter api.Submitter) {
	k.submitters = append(k.submitters, submitter)
}

// Run executes the main loop, which invokes scheduler plugins and binds pods to the selected nodes.
// This method blocks until ctx is done.
func (k *KubeSim) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.L.Debugf("Clock %s", k.clock.String())

			if err := k.submit(k.clock); err != nil {
				return err
			}

			if err := k.schedule(k.clock); err != nil {
				return err
			}

			k.clock = k.clock.Add(time.Duration(k.tick) * time.Second)
		}
	}
}

// List implements "k8s.io/pkg/scheduler/algorithm".NodeLister
func (k *KubeSim) List() ([]*v1.Node, error) {
	nodes := make([]*v1.Node, 0, len(k.nodes))
	for _, node := range k.nodes {
		nodes = append(nodes, node.ToV1())
	}
	return nodes, nil
}

func (k *KubeSim) submit(clock clock.Clock) error {
	nodes, _ := k.List()

	for _, submitter := range k.submitters {
		pods, err := submitter.Submit(clock, nodes)
		if err != nil {
			return err
		}

		for _, pod := range pods {
			pod.CreationTimestamp = clock.ToMetaV1()

			log.L.Tracef("Submit %v", pod)
			log.L.Debugf("Submit %q", pod.Name)

			k.podQueue.Push(pod)
		}
	}

	return nil
}

func (k *KubeSim) schedule(clock clock.Clock) error {
	nodeInfoMap := map[string]*nodeinfo.NodeInfo{}
	for name, node := range k.nodes {
		nodeInfoMap[name] = node.ToNodeInfo(clock)
	}

	results, err := k.scheduler.Schedule(k.podQueue, k, nodeInfoMap)
	if err != nil {
		return err
	}

	for _, result := range results {
		nodeName := result.Result.SuggestedHost
		node, ok := k.nodes[nodeName]
		if !ok {
			return errors.Errorf("No node named %q", nodeName)
		}

		if err := node.CreatePod(clock, result.Pod); err != nil {
			return err
		}
	}

	return nil
}

// readConfig reads and parses a config from the path (excluding file extension).
func readConfig(path string) (*config.Config, error) {
	viper.SetConfigName(path)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	log.G(context.TODO()).Debugf("Using config file %s", viper.ConfigFileUsed())

	var conf = config.Config{
		LogLevel:   "info",
		Tick:       10,
		StartClock: "",
		// APIPort:     10250,
		// MetricsPort: 10255,
		Cluster: config.ClusterConfig{Nodes: []config.NodeConfig{}},
	}

	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func configLog(logLevel string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return strongerrors.InvalidArgument(errors.Errorf("%s: log level %q not supported", err.Error(), level))
	}
	logrus.SetLevel(level)

	logger := log.L
	log.L = logger

	return nil
}
