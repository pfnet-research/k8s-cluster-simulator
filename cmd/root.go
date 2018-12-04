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

	"github.com/ordovicia/kubernetes-simulator/log"
	"github.com/ordovicia/kubernetes-simulator/sim"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Long:  "FIXME: kubernetes-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		_ = ctx

		clock := sim.NewTime(time.Now())

		config, err := initConfig(configPath)
		if err != nil {
			log.L.WithError(err).Fatal("Error building node config")
		}
		nodes := [](*sim.Node){}

		for _, config := range config.Cluster.Nodes {
			log.L.Debugf("NodeConfig: %+v", config)

			nodeConfig, err := buildNodeConfig(config)
			if err != nil {
				log.L.WithError(err).Fatal("Error building node config")
			}

			node := sim.NewNode(*nodeConfig)
			nodes = append(nodes, &node)
			log.L.Infof("Node %q created", nodeConfig.Name)
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
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (exclusing file extension)")
}

func initConfig(path string) (*Config, error) {
	viper.SetConfigName(path)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	} else {
		log.G(context.TODO()).Debugf("Using config file %s", viper.ConfigFileUsed())
	}

	var config = Config{
		Cluster:     ClusterConfig{Nodes: []NodeConfig{}},
		APIPort:     10250,
		MetricsPort: 10255,
		LogLevel:    "info",
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, strongerrors.InvalidArgument(errors.Errorf("log level %q not supported", level))
	}
	logrus.SetLevel(level)

	logger := log.L
	log.L = logger

	logger.Debugf("Config: %+v", config)

	return &config, nil
}
