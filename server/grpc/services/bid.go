package services

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	v1 "xrf197ilz35aq2/gen/go/service/v1"
	"xrf197ilz35aq2/internal/exchange"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
)

type BidService struct {
	Log            slog.Logger
	BidCacheClient redis.BidCache
	BidRepo        postgres.BidRepository
	SessionRepo    postgres.SessionRepository

	v1.UnimplementedBidServiceServer
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
		Bid: &v1.BidResponse{
			AssetId:   bid.AssetId,
			SessionId: activeSession.Id,
			LastUntil: request.LastUntil,
			Amount:    float32(bid.Amount),
			Quantity:  float32(bid.Quantity),
		},
	}, nil
}

func (srv *BidService) GetUserBid(ctx context.Context, request *v1.GetUserBidRequest) (*v1.GetUserBidResponse, error) {
	if request.AssetId == "" {
		return nil, fmt.Errorf("assetId is required")
	}
	if request.Offset < 0 {
		return nil, fmt.Errorf("offset must be greater than or equal to 0")
	}
	if request.Limit < 0 || request.Limit > 100 {
		return nil, fmt.Errorf("limit must be less than or equal to 100")
	}

	bids, err := srv.BidRepo.FetchBidsByUserFp(ctx, request.Offset, request.Limit, request.UserFp)
	if err != nil {
		return nil, err
	}
	var bidResponses []*v1.BidResponse
	for _, bid := range bids {
		bidResponses = append(bidResponses, &v1.BidResponse{
			AssetId:   bid.AssetId,
			SessionId: bid.SessionId,
			Amount:    float32(bid.Amount),
			Quantity:  float32(bid.Quantity),
			LastUntil: timestamppb.New(bid.LastUntil),
		})
	}
	return &v1.GetUserBidResponse{
		Bids:         bidResponses,
		Offset:       request.Offset,
		RowCount:     int64(len(bids)),
		TotalResults: 0, // TODO
	}, nil
}

func NewBidService(log slog.Logger, bidCache redis.BidCache) *BidService {
	return &BidService{
		Log:            log,
		BidCacheClient: bidCache,
	}
}
