package cmd

import (
	"k8s.io/api/core/v1"
)

// Config represents a simulator config by user
type Config struct {
	Cluster     ClusterConfig
	APIPort     int
	MetricsPort int
	LogLevel    string
	Taint       TaintConfig
}

type ClusterConfig struct {
	Nodes []NodeConfig
}

type NodeConfig struct {
	Name            string
	Capacity        map[v1.ResourceName]string
	OperatingSystem string
}

type TaintConfig struct {
	Key    string
	Value  string
	Effect string
}
