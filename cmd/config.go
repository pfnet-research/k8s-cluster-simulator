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
}

type ClusterConfig struct {
	Nodes []NodeConfig
}

type NodeConfig struct {
	Name     string
	Capacity map[v1.ResourceName]string
	Labels   map[string]string
	Taints   []TaintConfig
}

type TaintConfig struct {
	Key    string // TODO: force constraints
	Value  string
	Effect string
}
