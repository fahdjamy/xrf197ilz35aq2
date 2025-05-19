package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"xrf197ilz35aq2/gen/go/service/v1"
)

type bidService struct {
	log slog.Logger
	v1.UnimplementedBidServiceServer
}

func (srv *bidService) CreateBid(ctx context.Context, request *v1.CreateBidRequest) (*v1.CreateBidResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (srv *bidService) mustEmbedUnimplementedBidServiceServer() {
	//TODO implement me
	panic("implement me")
}

func NewGRPCSrv(log slog.Logger) (*grpc.Server, error) {
	// 1. Create a gRPC server object
	// Pass in server options here, like interceptors, TLS credentials, etc.
	grpcServer := grpc.NewServer()

	// 2. Register your service implementation with the gRPC server.
	v1.RegisterBidServiceServer(grpcServer, &bidService{
		log: log,
	})

	// 3. Optional: Register gRPC server reflection.
	// This allows gRPC clients (like grpcurl or a GUI client) to query what services and methods are available on
	// the server without needing the .proto file. Useful for debugging and exploration.
	reflection.Register(grpcServer)

	return grpcServer, nil
}
