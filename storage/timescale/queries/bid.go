package queries

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"xrf197ilz35aq2/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const bidRecordTableName = "bid_records"

type BidScanError struct {
	Err       error
	SkipCount int64
}

func (b *BidScanError) Error() string {
	return fmt.Sprintf("error scanning bid record :: err=%s, skipCount=%d", b.Err, b.SkipCount)
}

type BidTSQuerier interface {
	SaveBid(ctx context.Context, bid domain.Bid) (bool, error)
	BatchSave(ctx context.Context, bids []domain.Bid) (int64, error)
	FindBidsInTimeRange(ctx context.Context, startTime time.Time, endTime time.Time) ([]domain.Bid, error)
}

type bidTSQuerier struct {
	db  *pgxpool.Pool
	log slog.Logger
}

func (querier *bidTSQuerier) SaveBid(ctx context.Context, bid domain.Bid) (bool, error) {
	insertSQL := `
INSERT INTO bid_records
    (id, is_accepted, asset_id, bidder_fp, seller_fp, bid_time, session_id, amount, quantity, expiration_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
`
	result, err := querier.db.Exec(ctx, insertSQL, bid.Id,
		bid.AssetId, bid.UserFp, bid.AssetOwner,
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

func (querier *bidTSQuerier) FindBidsInTimeRange(ctx context.Context, startTime time.Time, endTime time.Time) ([]domain.Bid, error) {
	if startTime.After(endTime) {
		return nil, errors.New("start time must be before end time")
	}
	selectSQL := `
SELECT bid_id, symbol, is_accepted, bid_time, asset_id, bidder_fp, seller_fp, quantity,
       session_id, amount, quantity, expiration_time
	FROM
	    bid_records
	WHERE bid_time >= $1 AND bid_time <= $2
`
	rows, err := querier.db.Query(ctx, selectSQL, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("error querying bids in time range :: err=%w", err)
	}
	defer rows.Close()

	var bids []domain.Bid

	bidScanError := &BidScanError{}
	for rows.Next() {
		var bid domain.Bid
		err := rows.Scan(&bid.Id,
			&bid.Accepted,
			&bid.Timestamp,
			&bid.AssetId,
			&bid.UserFp,
			&bid.AssetOwner,
			&bid.Quantity,
			&bid.SessionId,
			&bid.Amount,
			&bid.LastUntil)
		if err != nil {
			bidScanError.Err = err // ⚠️!!IMPORTANT!! this will always overwrite the bidScanError error
			bidScanError.SkipCount++
			// log error encountered while scanning a bid
			querier.log.Error("error scanning bid record", "err", err)
			continue // skip this bid and continue to the next one
		}
		bids = append(bids, bid)
	}
	if bidScanError.Err != nil {
		return bids, bidScanError // return the error encountered while scanning bids and the successfully scanned bids
	}
	if err := rows.Err(); err != nil {
		return bids, fmt.Errorf("error scanning bid records :: err=%w", err) // return any other error encountered
	}
	return bids, nil // return the list of bids if no error was encountered
}

func NewBidTSQuerier(db *pgxpool.Pool, log slog.Logger) BidTSQuerier {
	return &bidTSQuerier{
		db:  db,
		log: log,
	}
}
