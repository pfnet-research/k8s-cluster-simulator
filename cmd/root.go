// Copyright © 2017 The virtual-kubelet authors
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

// Modification copyright @ 2018 <Name> <E-mail>

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
	corev1 "k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-scheduler-simulator/log"
)

var configFile string
var config = Config{
	Cluster:     ClusterConfig{Nodes: []NodeConfig{}},
	APIPort:     10250,
	MetricsPort: 10255,
	LogLevel:    "info",
	Taint: TaintConfig{
		Key:    "k8s-scheduler-simulator.io/kubelet",
		Value:  "simulator",
		Effect: "NoSchedule",
	},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "k8s-scheduler-simulator",
	Short: "k8s-scheduler-simulator provides a virtual kubernetes cluster interface for your kubernetes scheduler.",
	Long: `FIXME: virtual-kubelet implements the Kubelet interface with a pluggable
backend implementation allowing users to create kubernetes nodes without running the kubelet.
This allows users to schedule kubernetes workloads on nodes that aren't running Kubernetes.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, cancel := context.WithCancel(context.Background())

		for _, node := range config.Cluster.Nodes {
			log.L.Infof("node %q started", node.Name)
			time.Sleep(1 * time.Second)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sig
			cancel()
		}()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.G(context.TODO()).WithError(err).Fatal("Error executing root command")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (excluding file extension)")
}

// initConfig reads the config file.
func initConfig() {
	// TODO: Do not try to read config when 'help' or 'version' subcommand is provided

	viper.SetConfigName(configFile)
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

	taint, err := buildTaint(config.Taint)
	if err != nil {
		logger.WithError(err).Fatal("Error building taint")
	}
	_ = taint // TODO

	logger.Debugf("Config %+v", config)
}

// buildTaint builds a taint with the provided config.
func buildTaint(config TaintConfig) (*corev1.Taint, error) {
	var effect corev1.TaintEffect
	switch config.Effect {
	case "NoSchedule":
		effect = corev1.TaintEffectNoSchedule
	case "NoExecute":
		effect = corev1.TaintEffectNoExecute
	case "PreferNoSchedule":
		effect = corev1.TaintEffectPreferNoSchedule
	default:
		return nil, strongerrors.InvalidArgument(errors.Errorf("taint effect %q is not supported", config.Effect))
	}

	return &corev1.Taint{
		Key:    config.Key,
		Value:  config.Value,
		Effect: effect,
	}, nil
}
