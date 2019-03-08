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
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	"github.com/ordovicia/kubernetes-simulator/kubesim/scheduler"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// KubeSim represents a kubernetes cluster simulator.
type KubeSim struct {
	tick  time.Duration
	clock clock.Clock

	nodes    map[string]*node.Node
	podQueue queue.PodQueue

	submitters []api.Submitter
	scheduler  scheduler.Scheduler

	metricsWriters []metrics.Writer
	metricsTick    time.Duration
}

// NewKubeSim creates a new KubeSim with the given config, queue, and scheduler.
func NewKubeSim(conf *config.Config, queue queue.PodQueue, sched scheduler.Scheduler) (*KubeSim, error) {
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

		nodes:    nodes,
		podQueue: queue,

		submitters: []api.Submitter{},
		scheduler:  sched,

		metricsTick:    time.Duration(metricsTick) * time.Second,
		metricsWriters: metricsWriters,
	}, nil
}

// NewKubeSimFromConfigPath creates a new KubeSim with config from confPath (excluding file path),
// queue, and scheduler.
func NewKubeSimFromConfigPath(confPath string, queue queue.PodQueue, sched scheduler.Scheduler) (*KubeSim, error) {
	conf, err := readConfig(confPath)
	if err != nil {
		return nil, errors.Errorf("Error reading config: %s", err.Error())
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
	preMetricsClock := k.clock
	met, err := metrics.BuildMetrics(k.clock, k.nodes, k.podQueue)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.L.Debugf("Clock %s", k.clock.ToRFC3339())

			if err = k.submit(met); err != nil {
				return err
			}

			if err = k.schedule(); err != nil {
				return err
			}

			met, err = metrics.BuildMetrics(k.clock, k.nodes, k.podQueue)
			if err != nil {
				return err
			}

			if k.clock.Sub(preMetricsClock) > k.metricsTick {
				preMetricsClock = k.clock
				if err = k.writeMetrics(met); err != nil {
					return err
				}
			}

			k.clock = k.clock.Add(k.tick)
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
	level, err := log.ParseLevel(logLevel) // if logLevel == "", level <- info
	if err != nil {
		return strongerrors.InvalidArgument(errors.Errorf("Log level %q not supported: %s", level, err.Error()))
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
	for _, nodeConf := range conf.Cluster.Nodes {
		log.L.Debugf("Node config %+v", nodeConf)

		nodeV1, err := config.BuildNode(nodeConf, conf.StartClock)
		if err != nil {
			return map[string]*node.Node{}, errors.Errorf("Error building node config: %s", err.Error())
		}

		n := node.NewNode(nodeV1)
		nodes[nodeV1.Name] = &n

		log.L.Debugf("Node %s created", nodeV1.Name)
	}

	return nodes, nil
}

func buildMetricsWriters(conf *config.Config) ([]metrics.Writer, error) {
	writers := []metrics.Writer{}

	fileWriter, err := config.BuildMetricsFile(conf.MetricsFile)
	if err != nil {
		return []metrics.Writer{}, err
	}
	if fileWriter != nil {
		log.L.Infof("Metrics and log written to %s", fileWriter.FileName())
		writers = append(writers, fileWriter)
	}

	stdoutWriter, err := config.BuildMetricsStdout(conf.MetricsStdout)
	if err != nil {
		return []metrics.Writer{}, err
	}
	if stdoutWriter != nil {
		log.L.Info("Metrics and log written to Stdout")
		writers = append(writers, stdoutWriter)
	}

	return writers, nil
}

func (k *KubeSim) submit(metrics metrics.Metrics) error {
	if len(k.submitters) == 0 {
		return nil
	}

	for _, submitter := range k.submitters {
		events, err := submitter.Submit(k.clock, k, metrics)
		if err != nil {
			return err
		}

		for _, e := range events {
			if submitted, ok := e.(*api.SubmitEvent); ok {
				pod := submitted.Pod
				pod.CreationTimestamp = k.clock.ToMetaV1()
				pod.Status.Phase = v1.PodPending

				log.L.Tracef("Submit %v", pod)

				key, err := util.PodKey(pod)
				if err != nil {
					return err
				}
				log.L.Debugf("Submit %s", key)

				k.podQueue.Push(pod)
			} else if deleted, ok := e.(*api.DeleteEvent); ok {
				deletedFromQueue, err := k.podQueue.Delete(deleted.PodNamespace, deleted.PodName)
				if err != nil {
					return err
				}

				if !deletedFromQueue {
					// TODO
				}
			} else {
				panic("Unknown submitter event")
			}
		}
	}

	return nil
}

func (k *KubeSim) schedule() error {
	nodeInfoMap := map[string]*nodeinfo.NodeInfo{}
	for name, node := range k.nodes {
		nodeInfoMap[name] = node.ToNodeInfo(k.clock)
	}

	events, err := k.scheduler.Schedule(k.clock, k.podQueue, k, nodeInfoMap)
	if err != nil {
		return err
	}

	for _, e := range events {
		if bind, ok := e.(*scheduler.BindEvent); ok {
			nodeName := bind.ScheduleResult.SuggestedHost
			node, ok := k.nodes[nodeName]

			if ok {
				bind.Pod.Spec.NodeName = nodeName
			} else {
				return errors.Errorf("No node named %q", nodeName)
			}

			if err := node.BindPod(k.clock, bind.Pod); err != nil {
				return err
			}
		} else if delete, ok := e.(*scheduler.DeleteEvent); ok {
			_ = delete
			panic("Unimplemented")
		} else {
			panic("Unknown scheduler event")
		}
	}

	return nil
}

func (k *KubeSim) writeMetrics(met metrics.Metrics) error {
	for _, writer := range k.metricsWriters {
		if err := writer.Write(met); err != nil {
			return err
		}
	}

	return nil
}
