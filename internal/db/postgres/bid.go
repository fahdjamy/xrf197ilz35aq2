package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/db/postgres/dao"
	"xrf197ilz35aq2/internal/exchange"
)

type BidRepository interface {
	BatchCreateBids(ctx context.Context, bids []exchange.BidRequest, userFp string, sessionId int64) (int64, error)
	CreateBid(ctx context.Context, request exchange.BidRequest, userFp string, sessionId int64) (*domain.Bid, error)
	CreateBidsCopyFrom(ctx context.Context, bids []exchange.BidRequest, userFp string, sessionId int64) (int64, error)
}

type bidRepository struct {
	log    slog.Logger
	dbPool *pgx.Conn
}

func (repo *bidRepository) CreateBid(ctx context.Context, request exchange.BidRequest, userFp string, sessionId int64) (*domain.Bid, error) {
	newBid, err := domain.NewBid(userFp, request.Amount, request.AssetId, request.LastUntil, sessionId)
	if err != nil {
		return nil, err
	}
	sql := `
INSERT INTO bid (id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`
	var id int64
	err = repo.dbPool.QueryRow(ctx, sql,
		newBid.Amount,
		newBid.AssetId,
		newBid.Status,
		newBid.Accepted,
		newBid.UserFp,
		newBid.PlacedAt,
		newBid.LastUntil,
		newBid.SessionId,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	newBid.Id = id
	return newBid, nil
}

func (repo *bidRepository) BatchCreateBids(ctx context.Context, bids []exchange.BidRequest, userFp string, sessionId int64) (int64, error) {
	repo.log.Info(fmt.Sprintf("batch creating bids using, rowLen=%d", len(bids)))
	tx, err := repo.dbPool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("error starting transaction for batch create bids: %w", err)
	}
	defer tx.Rollback(ctx) //Rollback on error

	batch := &pgx.Batch{}
	for _, bid := range bids {
		newBid, err := domain.NewBid(userFp, bid.Amount, bid.AssetId, bid.LastUntil, sessionId)
		if err != nil {
			return 0, err
		}
		batch.Queue(`
INSERT INTO bid (id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			newBid.Id,
			newBid.Amount,
			newBid.AssetId,
			newBid.Status,
			newBid.Accepted,
			newBid.UserFp,
			newBid.PlacedAt,
			newBid.LastUntil,
			newBid.SessionId)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(bids); i++ {
		_, err := results.Exec()
		if err != nil {
			return 0, fmt.Errorf("error executing batch query %d: %w", i, err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("error committing batch create bids: %w", err)
	}
	return int64(len(bids)), nil
}

func (repo *bidRepository) CreateBidsCopyFrom(ctx context.Context, bids []exchange.BidRequest, userFp string, sessionId int64) (int64, error) {
	// The pgx.CopyFrom method provides a highly efficient way to bulk load data into a PostgresSQL table by leveraging the PostgresSQL COPY protocol
	// This method is significantly faster than executing individual INSERT statements or even using batched inserts for large datasets
	repo.log.Info(fmt.Sprintf("creating bulk bids using CopyFrom, rowLen=%d", len(bids)))
	rowSrc := pgx.CopyFromSlice(len(bids), func(i int) ([]interface{}, error) {
		newBid, err := domain.NewBid(userFp, bids[i].Amount, bids[i].AssetId, bids[i].LastUntil, sessionId)
		if err != nil {
			return nil, err
		}
		return []any{
			newBid.Id,
			newBid.Accepted,
			newBid.Status,
			newBid.AssetId,
			newBid.Amount,
			newBid.UserFp,
			newBid.SessionId,
			newBid.LastUntil,
			newBid.PlacedAt,
		}, nil
	})
	columnNames := dao.GetBidColumnName()
	count, err := repo.dbPool.CopyFrom(ctx, pgx.Identifier{dao.BidTableName}, columnNames, rowSrc)
	if err != nil {
		return 0, fmt.Errorf("error bulk copying/creating bid rows: %w", err)
	}
	return count, nil
}

func NewBidService(dbPool *pgx.Conn, log slog.Logger) BidRepository {
	return &bidRepository{
		dbPool: dbPool,
		log:    log,
	}
}
