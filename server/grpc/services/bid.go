package services

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"time"
	v1 "xrf197ilz35aq2/gen/go/service/v1"
	"xrf197ilz35aq2/internal/exchange"
	"xrf197ilz35aq2/server/socket"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
)

type bidService struct {
	hub            *socket.Hub
	Log            slog.Logger
	BidCacheClient redis.BidCache
	BidRepo        postgres.BidRepository
	SessionRepo    postgres.SessionRepository

	v1.UnimplementedBidServiceServer
}

func (srv *bidService) CreateBid(ctx context.Context, request *v1.CreateBidRequest) (*v1.CreateBidResponse, error) {
	// Extract metadata from the incoming context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		srv.Log.Error("Error: Missing metadata in context for create bid")
		// No metadata was sent at all, which is unusual for gRPC calls since gRPC often adds its own internal metadata.
		// Decide how to handle this; often if specific metadata is required.
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	// 2. Check for the required header "x-auth-token"
	// Header keys are conventionally lowercase in metadata.MD
	userFPValues := md.Get("x-rfz-user")
	if len(userFPValues) == 0 {
		srv.Log.Error("Error: Empty user fingerprint header in context for create bid")
		return nil, status.Errorf(codes.Unauthenticated, "missing user header")
	}

	// 3. Use the header value (taking the first one if multiple is sent)
	userFp := userFPValues[0]
	if userFp == "" { // Or perform more specific validation
		return nil, status.Errorf(codes.Unauthenticated, "user fingerprint is empty in header")
	}

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
		return nil, status.Errorf(codes.Internal, "failed to save bid")
	}
	return &v1.CreateBidResponse{
		Bid: &v1.BidResponse{
			BidId:     bid.Id,
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
		return nil, status.Errorf(codes.InvalidArgument, "assetId is required")
	}
	if request.Offset < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "offset must be greater than or equal to 0")
	}
	if request.Limit < 0 || request.Limit > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "limit must be less than or equal to 100")
	}

	bids, err := srv.BidRepo.FetchBidsByUserFp(ctx, request.Offset, request.Limit, request.UserFp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch bids")
	}
	var bidResponses []*v1.BidResponse
	for _, bid := range bids {
		bidResponses = append(bidResponses, &v1.BidResponse{
			BidId:     bid.Id,
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
		return status.Errorf(codes.InvalidArgument, "assetId is required")
	}
	if req.Offset < 0 {
		return status.Errorf(codes.InvalidArgument, "offset must be greater than or equal to 0")
	}
	if req.Limit < 0 || req.Limit > 200 {
		return status.Errorf(codes.InvalidArgument, "limit must be less than or equal to 200")
	}

	limit := req.Limit
	offset := req.Offset
	activeSession, err := srv.SessionRepo.FindActiveSession(srvStream.Context(), req.AssetId)
	if err != nil {
		return err
	}
	if activeSession == nil {
		return status.Errorf(codes.NotFound, "no active session found for assetId=%s", req.AssetId)
	}
	srv.Log.Info("streaming open bids", "assetId", req.AssetId, "sessionId", activeSession.Id)

	for {
		// Check if the client context is canceled (e.g., a client disconnected)
		if err = srvStream.Context().Err(); err != nil {
			return status.Errorf(codes.Canceled, "stream canceled")
		}

		pgCtx, cancelPgCtx := context.WithTimeout(srvStream.Context(), 10*time.Second)
		bids, err := srv.BidRepo.FetchBidsByAssetIdAndSessionId(pgCtx, offset, limit, req.AssetId, activeSession.Id)
		if err != nil {
			cancelPgCtx()
			return status.Errorf(codes.Internal, "failed to fetch bids")
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
				BidId:     bid.Id,
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
			return status.Errorf(codes.Internal, "failed to send bid response")
		}
	}
}

func NewBidService(log slog.Logger, bidCache redis.BidCache, BidRepo postgres.BidRepository, hub *socket.Hub) v1.BidServiceServer {
	return &bidService{
		hub:            hub,
		Log:            log,
		BidRepo:        BidRepo,
		BidCacheClient: bidCache,
	}
}
