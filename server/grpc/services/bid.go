package services

import (
	"context"
	"log/slog"
	v1 "xrf197ilz35aq2/gen/go/service/v1"
	"xrf197ilz35aq2/internal/exchange"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
)

type BidService struct {
	Log            slog.Logger
	BidCacheClient redis.BidCache
	v1.UnimplementedBidServiceServer
	SessionRepo postgres.SessionRepository
}

func (srv *BidService) CreateBid(ctx context.Context, request *v1.CreateBidRequest) (*v1.CreateBidResponse, error) {

	userFp := "should be fetched from auth token"

	activeSession, err := srv.SessionRepo.FindActiveSession(ctx, request.AssetId)
	if err != nil {
		return nil, err
	}
	srv.Log.Info("placing bid", "assetId", request.AssetId, "sessionId", activeSession.Id)

	bidRequest := exchange.BidRequest{
		UserFp:    userFp,
		AssetId:   request.AssetId,
		Amount:    float64(request.Amount),
		LastUntil: request.LastUntil.AsTime(),
	}

	bid, err := srv.BidCacheClient.SaveBid(ctx, bidRequest, activeSession.Id)
	if err != nil {
		return nil, err
	}
	return &v1.CreateBidResponse{
		Created:    true,
		BidId:      bid.Id,
		IsAccepted: bid.Accepted,
		SessionId:  activeSession.Id,
	}, nil
}

func NewBidService(log slog.Logger, bidCache redis.BidCache) *BidService {
	return &BidService{
		Log:            log,
		BidCacheClient: bidCache,
	}
}
