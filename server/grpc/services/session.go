package services

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	v1 "xrf197ilz35aq2/gen/go/service/session/v1"
	"xrf197ilz35aq2/internal/exchange"
	"xrf197ilz35aq2/storage/postgres"
)

type sessionService struct {
	log         slog.Logger
	sessionRepo postgres.SessionRepository

	v1.UnimplementedSessionServiceServer
}

func (srvc *sessionService) CreateSession(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	sessionName := ""
	if req.Name != nil {
		sessionName = *req.Name
	}
	sessionReq := exchange.NewSessionRequest{
		AssetId:            req.AssetId,
		Name:               sessionName,
		Type:               req.AuctionType,
		AutoExecute:        req.AutoExecute,
		EndTime:            req.EndTime.AsTime(),
		StartTime:          req.StartTime.AsTime(),
		ReservePrice:       float64(req.ReservePrice),
		BidIncrementAmount: float64(req.BidIncrementAmount),
	}

	newSession, err := domain.NewSession(sessionReq, "")
	if err != nil {
		return nil, err
	}

	createdSessionId, err := srvc.sessionRepo.Create(ctx, newSession)

	if err != nil {
		return nil, err
	}
	return &v1.CreateSessionResponse{
		Session: &v1.SessionResponse{
			SessionId:          createdSessionId,
			Name:               &newSession.Name,
			Status:             newSession.Status,
			AuctionType:        newSession.ActionType,
			AutoExecute:        newSession.AutoExecute,
			ReservePrice:       float32(newSession.ReservePrice),
			EndTime:            timestamppb.New(newSession.EndTime),
			StartTime:          timestamppb.New(newSession.StartTime),
			CreatedAt:          timestamppb.New(newSession.CreatedAt),
			CurrentHighestBid:  float32(newSession.CurrentHighestBid),
			BidIncrementAmount: float32(newSession.BidIncrementAmount),
		},
	}, nil
}

func (srvc *sessionService) GetActiveAssetSession(ctx context.Context, req *v1.GetActiveAssetSessionRequest) (*v1.GetActiveAssetSessionResponse, error) {
	activeSession, err := srvc.sessionRepo.FindActiveSession(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	if activeSession == nil {
		return nil, fmt.Errorf("no active session found for assetId=%s", req.AssetId)
	}

	return &v1.GetActiveAssetSessionResponse{
		Session: &v1.SessionResponse{
			SessionId:          activeSession.Id,
			Name:               &activeSession.Name,
			Status:             activeSession.Status,
			AuctionType:        activeSession.ActionType,
			AutoExecute:        activeSession.AutoExecute,
			ReservePrice:       float32(activeSession.ReservePrice),
			EndTime:            timestamppb.New(activeSession.EndTime),
			StartTime:          timestamppb.New(activeSession.StartTime),
			CreatedAt:          timestamppb.New(activeSession.CreatedAt),
			CurrentHighestBid:  float32(activeSession.CurrentHighestBid),
			BidIncrementAmount: float32(activeSession.BidIncrementAmount),
		},
	}, nil
}

func NewSessionServiceServer(log slog.Logger, sessionRepo postgres.SessionRepository) v1.SessionServiceServer {
	return &sessionService{
		log:         log,
		sessionRepo: sessionRepo,
	}
}
