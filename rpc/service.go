package rpc

import (
	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Microservice represents a gRPC server that implements the ServiceServer interface and wraps a map matcher engine
type Microservice struct {
	protos_pb.ServiceServer
	matcher *horizon.MapMatcher
}

// NewMicroserice creates a new gRPC server instance with the provided matcher and reflection option.
func NewMicroserice(matcher *horizon.MapMatcher, reflect bool) (*grpc.Server, error) {
	grpcInstance := grpc.NewServer()

	server := Microservice{
		matcher: matcher,
	}
	protos_pb.RegisterServiceServer(
		grpcInstance,
		&server,
	)
	if reflect {
		reflection.Register(grpcInstance)
	}
	return grpcInstance, nil
}
