package queries

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
)

const bidRecordTableName = "bid_records"

type BidTSQuerier interface {
	SaveBid(ctx context.Context, bid domain.Bid) (bool, error)
	BatchSave(ctx context.Context, bids []domain.Bid) (int64, error)
	GetBidById(ctx context.Context, bidId string) (domain.Bid, error)
	BatchSaveBid(ctx context.Context, bids []domain.Bid) (int64, error)
}

type bidTSQuerier struct {
	db  *pgxpool.Pool
	log slog.Logger
}

func (querier *bidTSQuerier) SaveBid(ctx context.Context, bid domain.Bid) (bool, error) {
	insertSQL := `
INSERT INTO bid_records
    (id, symbol, is_accepted, asset_id, bidder_fp, seller_fp, bid_time, session_id, amount, quantity, expiration_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
`
	result, err := querier.db.Exec(ctx, insertSQL, bid.Id,
		bid.Symbol, bid.AssetId, bid.UserFp, bid.AssetOwner,
		bid.Timestamp, bid.SessionId, bid.Amount,
		bid.Quantity, bid.LastUntil)
	if err != nil {
		return false, err
	}

	querier.log.Info("saved bid record to timescale", "rowsAffected", result.RowsAffected(), "bidId", bid.Id)
	return true, nil
}

func (querier *bidTSQuerier) BatchSave(ctx context.Context, bids []domain.Bid) (int64, error) {
	tx, err := querier.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("error starting transaction for batch save bids in TS-DB :: err=%w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			querier.log.Error("failed to rollback transaction for batch create bids", "err", err)
		}
	}(tx, ctx) //Rollback on error

	copyData := make([][]interface{}, len(bids))
	for i, bid := range bids {
		copyData[i] = []interface{}{
			bid.Id,
			bid.Symbol,
			bid.AssetId,
			bid.UserFp,
			bid.AssetOwner,
			bid.Timestamp,
			bid.SessionId,
			bid.Amount,
			bid.Quantity,
			bid.LastUntil,
		}
	}

	tableName := pgx.Identifier{bidRecordTableName}
	columnNames := []string{"id", "symbol", "asset_id", "bidder_fp", "seller_fp", "bid_time",
		"session_id", "amount", "quantity", "expiration_time"}

	count, err := querier.db.CopyFrom(ctx, tableName, columnNames, pgx.CopyFromRows(copyData))
	if err != nil {
		var pgErr *pgconn.PgError
		// Check for unique constraint violation (error code 23505)
		if ok := errors.As(err, &pgErr); ok && pgErr.Code == "23505" {
			// Specifically, check if it's our intended unique constraint
			if pgErr.ConstraintName == "bid_records_bid_id_unique" {
				querier.log.Warn("Warning: Unique constraint 'bid_records_bid_id_unique' violation during COPY to"+
					" bid_records (duplicate bid_id+bid_time)", "err", err)
				// This means a bid with the same ID and time already exists.
				// For historical data, this might be acceptable to skip.
			} else {
				// Another unique constraint was violated, or a different "23505" error.
				querier.log.Warn("Warning: A unique constraint violation occurred during COPY", "err", err)
			}
			// Decide if this is acceptable or requires full rollback.
			// For this example, we log and continue, assuming such duplicates are okay to skip.
		} else {
			return 0, fmt.Errorf("error copying data to timescale for batch save bids :: err=%w", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("error committing transaction for batch save bids :: err=%w", err)
	}
	return count, nil
}

func (querier *bidTSQuerier) GetBidById(ctx context.Context, bidId string) (domain.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (querier *bidTSQuerier) BatchSaveBid(ctx context.Context, bids []domain.Bid) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func NewBidTSQuerier(db *pgxpool.Pool, log slog.Logger) BidTSQuerier {
	return &bidTSQuerier{
		db:  db,
		log: log,
	}
}
