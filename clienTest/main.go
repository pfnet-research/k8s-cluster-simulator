package main

import (
	"log"

	"google.golang.org/grpc"

	clientPkg "simulator/pkg/client"
	pb "simulator/protos"
)

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

	clientPkg.SendMetric(clientPkg.Client, &metric)
}
