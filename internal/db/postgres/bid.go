package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/exchange"
)

type BidRepository interface {
	CreateBid(request exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) (*domain.Bid, error)
	BatchCreateBids(bids []exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) error
}

type bidRepository struct {
	log    slog.Logger
	dbPool *pgx.Conn
}

func (repo *bidRepository) CreateBid(request exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) (*domain.Bid, error) {
	newBid, err := domain.NewBid(userFp, request.Amount, request.AssetId, request.LastUntil, sessionId)
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
		newBid.Status,
		newBid.Accepted,
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

func (repo *bidRepository) BatchCreateBids(bids []exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) error {
	tx, err := repo.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction for batch create bids: %w", err)
	}
	defer tx.Rollback(ctx) //Rollback on error

	batch := &pgx.Batch{}
	for _, bid := range bids {
		newBid, err := domain.NewBid(userFp, bid.Amount, bid.AssetId, bid.LastUntil, sessionId)
		if err != nil {
			return err
		}
		batch.Queue(`
INSERT INTO bid (amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`,
			newBid.Amount,
			newBid.AssetId,
			newBid.Status,
			newBid.Accepted,
			newBid.UserFp,
			newBid.Placed,
			newBid.LastUntil,
			newBid.SessionId)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(bids); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("error executing batch query %d: %w", i, err)
		}
	}
	return tx.Commit(ctx)
}

func NewBidService(dbPool *pgx.Conn, log slog.Logger) BidRepository {
	return &bidRepository{
		dbPool: dbPool,
		log:    log,
	}
}
