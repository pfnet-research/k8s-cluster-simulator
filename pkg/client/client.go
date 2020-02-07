package client

import (
	pb "simulator/protos"
)

var Connect pb.SimRPCClient

func establishConnection() {
	address = "localhost:50051"

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	Connect = pb.NewSimRPCClient(conn)

	var metric pb.Metrics
	metric.clock = "test clock"
	metric.nodes = "test node"

}