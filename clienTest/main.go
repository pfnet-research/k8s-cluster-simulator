package main

import (
	"context"
	"log"
	"time"

	pb "simulator/protos"

	"google.golang.org/grpc"
)

var Connect pb.SimRPCClient

func main() {
	address := "localhost:50051"

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	Connect = pb.NewSimRPCClient(conn)

	var clock pb.Clock = pb.Clock{ClockMetrics_Key: "test clock"}
	var node pb.Nodes = pb.Nodes{NodesMetricsKey: "test node"}
	var pod pb.Pods = pb.Pods{PodsMetricsKey: "test pod"}
	var queue pb.Queue = pb.Queue{QueueMetricsKey: "test queue"}
	var metric pb.Metrics = pb.Metrics{
		Clock: &clock,
		Nodes: &node,
		Pods:  &pod,
		Queue: &queue}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := Connect.RecordMetrics(ctx, &metric)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %d", r.GetResult())
}
