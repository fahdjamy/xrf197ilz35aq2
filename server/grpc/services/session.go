package services

import (
	"context"
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

func NewSessionServiceServer(log slog.Logger, sessionRepo postgres.SessionRepository) v1.SessionServiceServer {
	return &sessionService{
		log:         log,
		sessionRepo: sessionRepo,
	}
}
