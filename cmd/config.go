package cmd

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
	Name     string
	Capacity CapacityConfig
}

type CapacityConfig struct {
	CPU    string
	Memory string
	GPU    string
	Pods   string
}

type TaintConfig struct {
	Key    string
	Value  string
	Effect string
}
