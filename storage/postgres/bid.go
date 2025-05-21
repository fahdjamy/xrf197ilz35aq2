package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/storage/postgres/dao"
)

type BidRepository interface {
	CreateBid(ctx context.Context, request domain.Bid) (string, error)
	BatchCreateBids(ctx context.Context, bids []domain.Bid) (int64, error)
	CreateBidsCopyFrom(ctx context.Context, bids []domain.Bid) (int64, error)
	FetchBidsByUserFp(ctx context.Context, offset int64, limit int64, userFp string) ([]domain.Bid, error)
	FetchBidsByAssetIdAndSessionId(ctx context.Context, offset int64, limit int64, assetId string, sessionId string) ([]domain.Bid, error)
}

type bidRepository struct {
	log    slog.Logger
	dbPool *pgxpool.Pool
}

func (repo *bidRepository) CreateBid(ctx context.Context, newBid domain.Bid) (string, error) {
	sql := `
INSERT INTO bid (id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`
	var id string
	err := repo.dbPool.QueryRow(ctx, sql,
		newBid.Amount,
		newBid.AssetId,
		newBid.Status,
		newBid.Accepted,
		newBid.UserFp,
		newBid.Timestamp,
		newBid.LastUntil,
		newBid.SessionId,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	newBid.Id = id
	return id, nil
}

func (repo *bidRepository) BatchCreateBids(ctx context.Context, bids []domain.Bid) (int64, error) {
	repo.log.Info(fmt.Sprintf("batch creating bids using, rowLen=%d", len(bids)))
	tx, err := repo.dbPool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("error starting transaction for batch create bids: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			repo.log.Error(fmt.Sprintf("error rolling back transaction for batch create bids: %w", err))
		}
	}(tx, ctx) //Rollback on error

	batch := &pgx.Batch{}
	for _, bid := range bids {
		batch.Queue(`
INSERT INTO bid (id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			bid.Id,
			bid.Amount,
			bid.AssetId,
			bid.Status,
			bid.Accepted,
			bid.UserFp,
			bid.Timestamp,
			bid.LastUntil,
			bid.SessionId)
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

func (repo *bidRepository) CreateBidsCopyFrom(ctx context.Context, bids []domain.Bid) (int64, error) {
	// The pgx.CopyFrom method provides a highly efficient way to bulk load data into a PostgresSQL table by leveraging the PostgresSQL COPY protocol
	// This method is significantly faster than executing individual INSERT statements or even using batched inserts for large datasets
	repo.log.Info(fmt.Sprintf("creating bulk bids using CopyFrom, rowLen=%d", len(bids)))
	rowSrc := pgx.CopyFromSlice(len(bids), func(i int) ([]interface{}, error) {
		bid := bids[i]
		return []any{
			bid.Id,
			bid.Accepted,
			bid.Status,
			bid.AssetId,
			bid.Amount,
			bid.UserFp,
			bid.SessionId,
			bid.LastUntil,
			bid.Timestamp,
		}, nil
	})
	columnNames := dao.GetBidColumnName()
	count, err := repo.dbPool.CopyFrom(ctx, pgx.Identifier{dao.BidTableName}, columnNames, rowSrc)
	if err != nil {
		return 0, fmt.Errorf("error bulk copying/creating bid rows: %w", err)
	}
	return count, nil
}

func (repo *bidRepository) FetchBidsByUserFp(ctx context.Context, offset int64, limit int64, userFp string) ([]domain.Bid, error) {
	sql := `
SELECT id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id
FROM bid
WHERE placed_by = $1
ORDER BY placed_at DESC
LIMIT $2 OFFSET $3`
	rows, err := repo.dbPool.Query(ctx, sql, userFp, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error fetching bids by user_fp: %w", err)
	}
	// Close the rows when we're done with them'
	defer rows.Close()
	var bids []domain.Bid
	rowScanError := &RowScanError{}
	for rows.Next() {
		var bid domain.Bid
		err = rows.Scan(
			&bid.Id,
			&bid.Amount,
			&bid.AssetId,
			&bid.Status,
			&bid.Accepted,
			&bid.UserFp,
			&bid.Timestamp,
			&bid.LastUntil,
			&bid.SessionId,
		)
		if err != nil {
			rowScanError.Err = err // ⚠️!!IMPORTANT!! this will always overwrite the rowScanError error
			rowScanError.SkipCount++
			repo.log.Error("error scanning bid record", "err", err) // log errors encountered while scanning
			continue                                                // skip this bid and continue to the next one
		}
		bids = append(bids, bid)
	}
	if rowScanError.Err != nil {
		return bids, rowScanError // return the error encountered while scanning bids and the successfully scanned bids
	}

	if err := rows.Err(); err != nil {
		return bids, fmt.Errorf("error scanning bid records: %w", err) // return any other error encountered
	}
	return bids, nil
}

func (repo *bidRepository) FetchBidsByAssetIdAndSessionId(ctx context.Context, offset int64, limit int64, assetId string, sessionId string) ([]domain.Bid, error) {
	sql := `
SELECT id, amount, asset_id, bid_status, accepted, placed_by, placed_at, last_until, session_id
FROM bid
WHERE asset_id = $1 AND session_id = $2
ORDER BY placed_at DESC
LIMIT $3 OFFSET $4`
	rows, err := repo.dbPool.Query(ctx, sql, assetId, sessionId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("error fetching bids by asset_id and session_id: %w", err)
	}
	// Close the rows when we're done with them'
	defer rows.Close()
	var bids []domain.Bid
	rowScanError := &RowScanError{}
	for rows.Next() {
		var bid domain.Bid
		err = rows.Scan(
			&bid.Id,
			&bid.Amount,
			&bid.AssetId,
			&bid.Status,
			&bid.Accepted,
			&bid.UserFp,
			&bid.Timestamp,
		)
		if err != nil {
			rowScanError.Err = err
			rowScanError.SkipCount++
			repo.log.Error("error scanning bid record", "err", err)
			continue
		}
		bids = append(bids, bid)
	}
	if rowScanError.Err != nil {
		return bids, rowScanError
	}

	if err := rows.Err(); err != nil {
		return bids, fmt.Errorf("error scanning bid records: %w", err)
	}
	return bids, nil
}

func NewBidRepo(dbPool *pgxpool.Pool, log slog.Logger) BidRepository {
	return &bidRepository{
		dbPool: dbPool,
		log:    log,
	}
}
