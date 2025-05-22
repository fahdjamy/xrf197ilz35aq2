package services

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"time"
	v1 "xrf197ilz35aq2/gen/go/service/v1"
	"xrf197ilz35aq2/internal/exchange"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
)

type bidService struct {
	Log            slog.Logger
	BidCacheClient redis.BidCache
	BidRepo        postgres.BidRepository
	SessionRepo    postgres.SessionRepository

	v1.UnimplementedBidServiceServer
}

func (srv *bidService) CreateBid(ctx context.Context, request *v1.CreateBidRequest) (*v1.CreateBidResponse, error) {

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

func (srv *bidService) GetUserBid(ctx context.Context, request *v1.GetUserBidRequest) (*v1.GetUserBidResponse, error) {
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
		TotalResults: 0, // TODO fetch actual total results
	}, nil
}

func (srv *bidService) StreamOpenBids(req *v1.StreamOpenBidsRequest, srvStream grpc.ServerStreamingServer[v1.StreamOpenBidsResponse]) error {
	if req.AssetId == "" {
		return fmt.Errorf("assetId is required")
	}
	if req.Offset < 0 {
		return fmt.Errorf("offset must be greater than or equal to 0")
	}
	if req.Limit < 0 || req.Limit > 200 {
		return fmt.Errorf("limit must be less than or equal to 100")
	}

	limit := req.Limit
	offset := req.Offset
	activeSession, err := srv.SessionRepo.FindActiveSession(srvStream.Context(), req.AssetId)
	if err != nil {
		return err
	}
	if activeSession == nil {
		return fmt.Errorf("no active session found for assetId=%s", req.AssetId)
	}
	srv.Log.Info("streaming open bids", "assetId", req.AssetId, "sessionId", activeSession.Id)

	for {
		// Check if the client context is canceled (e.g., a client disconnected)
		if err = srvStream.Context().Err(); err != nil {
			return err
		}

		pgCtx, cancelPgCtx := context.WithTimeout(srvStream.Context(), 10*time.Second)
		bids, err := srv.BidRepo.FetchBidsByAssetIdAndSessionId(pgCtx, offset, limit, req.AssetId, activeSession.Id)
		if err != nil {
			cancelPgCtx()
			return err
		}
		cancelPgCtx()

		// if there are no more bids, stop streaming and return
		// TODO: Should we sleep or exist?
		if len(bids) == 0 {
			return nil
		}

		// Send the response message over the stream.
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

		// update offset value
		offset += limit

		err = srvStream.Send(&v1.StreamOpenBidsResponse{
			Bids: bidResponses,
		})
		if err != nil {
			return err
		}
	}
}

func NewBidService(log slog.Logger, bidCache redis.BidCache, BidRepo postgres.BidRepository) v1.BidServiceServer {
	return &bidService{
		Log:            log,
		BidCacheClient: bidCache,
		BidRepo:        BidRepo,
	}
}
