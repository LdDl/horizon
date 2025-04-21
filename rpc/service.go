package rpc

import (
	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"google.golang.org/grpc"
)

type Microservice struct {
	protos_pb.ServiceServer
	matcher *horizon.MapMatcher
}

func NewMicroserice(matcher *horizon.MapMatcher) (*grpc.Server, error) {
	grpcInstance := grpc.NewServer()

	server := Microservice{
		matcher: matcher,
	}
	protos_pb.RegisterServiceServer(
		grpcInstance,
		&server,
	)
	// reflection.Register(grpcInstance)
	return grpcInstance, nil
}
