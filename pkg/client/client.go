package client

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/qrluo96/k8s-cluster-simulator/protos"
)

const (
	port = ":50051"
)

