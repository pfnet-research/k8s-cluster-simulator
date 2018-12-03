package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/ordovicia/kubernetes-simulator/log"
	"github.com/ordovicia/kubernetes-simulator/sim"
)

var configPath string
var config = Config{
	Cluster:     ClusterConfig{Nodes: []NodeConfig{}},
	APIPort:     10250,
	MetricsPort: 10255,
	LogLevel:    "info",
}
var nodeConfigs []sim.NodeConfig

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Long:  "FIXME: kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		_ = ctx

		clock := sim.NewTime(time.Now())
		nodes := []sim.Node{}

		for _, nodeConfig := range nodeConfigs {
			nodes = append(nodes, sim.NewNode(nodeConfig))
			log.L.Infof("node %q created", nodeConfig.Name)
		}

		// if err != nil {
		// 	log.L.WithError(err).Fatal("Error initializing virtual kubelet")
		// }

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			cancel()
		}()

		for _, node := range nodes {
			node.UpdateState(clock)
		}

		// if err := f.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
		// 	log.L.Fatal(err)
		// }
	},
}

// Execute executes the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (excluding file extension)")
}

func initConfig() {
	// TODO: Do not try to read config when 'help' or 'version' subcommand is provided

	viper.SetConfigName(configPath)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.G(context.TODO()).WithError(err).Fatal("Error reading config file")
	} else {
		log.G(context.TODO()).Debugf("Using config file %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.G(context.TODO()).WithError(err).Fatal("Error decoding config")
	}

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.G(context.TODO()).WithField("logLevel", config.LogLevel).Fatal("log level is not supported")
	}
	logrus.SetLevel(level)

	logger := log.L
	log.L = logger

	logger.Debugf("config: %+v", config)

	for _, nodeConfig := range config.Cluster.Nodes {
		capacity, err := buildCapacity(nodeConfig.Capacity)
		if err != nil {
			logger.WithError(err).Fatal("Error building capacity")
		}

		taints := []v1.Taint{}
		for _, taintConfig := range nodeConfig.Taints {
			taint, err := buildTaint(taintConfig)
			if err != nil {
				logger.WithError(err).Fatal("Error building taint")
			}
			taints = append(taints, *taint)
		}

		nodeConfig := sim.NodeConfig{
			Name:     nodeConfig.Name,
			Capacity: capacity,
			Labels:   nodeConfig.Labels,
			Taints:   taints,
		}

		nodeConfigs = append(nodeConfigs, nodeConfig)
	}

	logger.Debugf("nodeConfigs: %+v", nodeConfigs)
}

func buildCapacity(config map[v1.ResourceName]string) (v1.ResourceList, error) {
	resourceList := v1.ResourceList{}

	for key, value := range config {
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return nil, strongerrors.InvalidArgument(errors.Errorf("invalid %s value %q", key, value))
		}
		resourceList[key] = quantity
	}

	return resourceList, nil
}

func buildTaint(config TaintConfig) (*v1.Taint, error) {
	var effect v1.TaintEffect
	switch config.Effect {
	case "NoSchedule":
		effect = v1.TaintEffectNoSchedule
	case "NoExecute":
		effect = v1.TaintEffectNoExecute
	case "PreferNoSchedule":
		effect = v1.TaintEffectPreferNoSchedule
	default:
		return nil, strongerrors.InvalidArgument(errors.Errorf("taint effect %q is not supported", config.Effect))
	}

	return &v1.Taint{
		Key:    config.Key,
		Value:  config.Value,
		Effect: effect,
	}, nil
}
