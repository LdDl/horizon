package rpc

import (
	"context"

	"github.com/LdDl/horizon/rpc/protos_pb"
)

// GetSP Implement GetSP() to match interface
func (ts *Microservice) GetSP(ctx context.Context, in *protos_pb.SPRequest) (*protos_pb.SPResponse, error) {
	return &protos_pb.SPResponse{}, nil
}
