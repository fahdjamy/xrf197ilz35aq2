package queries

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
)

type BidTSQuerier interface {
	SaveBid(ctx context.Context, sellerFp string, bid domain.Bid) (bool, error)
	BatchSave(bids []domain.Bid) (int64, error)
	GetBidById(bidId string) (domain.Bid, error)
	BatchSaveBid(bids []domain.Bid) (int64, error)
}

type bidTSQuerier struct {
	db  *pgxpool.Pool
	log slog.Logger
}

func (querier *bidTSQuerier) SaveBid(ctx context.Context, sellerFp string, bid domain.Bid) (bool, error) {
	insertSQL := `
INSERT INTO bid_records
    (id, symbol, is_accepted, asset_id, bidder_fp, seller_fp, trade_time, session_id, amount, quantity, expiration_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
`
	result, err := querier.db.Exec(ctx, insertSQL, bid.Id,
		bid.Symbol, bid.AssetId, bid.UserFp, sellerFp,
		bid.PlacedAt, bid.SessionId, bid.Amount,
		bid.Quantity, bid.LastUntil)
	if err != nil {
		return false, err
	}

	querier.log.Info("saved bid record to timescale", "rowsAffected", result.RowsAffected(), "bidId", bid.Id)
	return true, nil
}

func (querier *bidTSQuerier) BatchSave(bids []domain.Bid) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (querier *bidTSQuerier) GetBidById(bidId string) (domain.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (querier *bidTSQuerier) BatchSaveBid(bids []domain.Bid) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func NewBidTSQuerier(db *pgxpool.Pool, log slog.Logger) BidTSQuerier {
	return &bidTSQuerier{
		db:  db,
		log: log,
	}
}
