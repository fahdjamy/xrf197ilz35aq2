package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"xrf197ilz35aq2/gen/go/service/v1"
	"xrf197ilz35aq2/server/grpc/services"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
)

func NewGRPCSrv(log slog.Logger, cacheClient redis.CacheClients, repos postgres.Repositories) (*grpc.Server, error) {
	// 1. Create a gRPC server object
	// Pass in server options here, like interceptors, TLS credentials, etc.
	grpcServer := grpc.NewServer()

	// 2. Register your service implementation with the gRPC server.
	v1.RegisterBidServiceServer(grpcServer, services.NewBidService(log, cacheClient.BidClient, repos.BidRepository))

	// 3. Optional: Register gRPC server reflection.
	// This allows gRPC clients (like grpcurl or a GUI client) to query what services and methods are available on
	// the server without needing the .proto file. Useful for debugging and exploration.
	reflection.Register(grpcServer)

	return grpcServer, nil
}
