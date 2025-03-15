package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/exchange"
)

type BidRepository interface {
	CreateBid(request exchange.BidRequest, userFp string, assetId string, ctx context.Context) (*domain.Bid, error)
}

type bidRepository struct {
	log    slog.Logger
	dbPool *pgx.Conn
}

func (repo *bidRepository) CreateBid(request exchange.BidRequest, userFp string, assetId string, ctx context.Context) (*domain.Bid, error) {
	newBid, err := domain.NewBid(userFp, request.Amount, assetId, request.LastUntil, request.SessionId)
	if err != nil {
		return nil, err
	}
	sql := `
INSERT INTO bid (amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`
	var id int64
	err = repo.dbPool.QueryRow(ctx, sql,
		newBid.Amount,
		newBid.AssetId,
		newBid.SessionId,
		newBid.Status,
		newBid.UserFp,
		newBid.Placed,
		newBid.LastUntil,
		newBid.SessionId,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	newBid.Id = id
	return newBid, nil
}

func NewBidService(dbPool *pgx.Conn, log slog.Logger) BidRepository {
	return &bidRepository{
		dbPool: dbPool,
		log:    log,
	}
}
