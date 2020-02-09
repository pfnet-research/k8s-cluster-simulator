package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"

	clientPkg "simulator/pkg/client"
	pb "simulator/protos"
)

// var Client pb.SimRPCClient

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func sendMetric(client pb.SimRPCClient, metric *pb.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := clientPkg.Client.RecordMetrics(ctx, metric)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %d", r.GetResult())
}

func initMetric(metric *pb.Metrics, clockStr string, nodeStr string) {
	var clock pb.Clock = pb.Clock{ClockMetrics_Key: clockStr}
	var node pb.Nodes = pb.Nodes{NodesMetricsKey: nodeStr}
	var pod pb.Pods = pb.Pods{PodsMetricsKey: "test pod"}
	var queue pb.Queue = pb.Queue{QueueMetricsKey: "test queue"}
	metric.Clock = &clock
	metric.Nodes = &node
	metric.Pods = &pod
	metric.Queue = &queue
}

func main() {
	address := "localhost:50051"

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	clientPkg.Client = pb.NewSimRPCClient(conn)

	var metric pb.Metrics
	clientPkg.InitMetric(&metric, "test clock", "test node")

	sendMetric(clientPkg.Client, &metric)
}
