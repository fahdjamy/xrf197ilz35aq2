package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
)

type SessionRepository interface {
	Create(session *domain.Session, ctx context.Context) (int64, error)
	FindById(sessionId string, ctx context.Context) (*domain.Session, error)
	FindByAssetId(assetId string, ctx context.Context) (*domain.Session, error)
}

type sessionRepository struct {
	log    slog.Logger
	dbPool *pgx.Conn
}

func (ses *sessionRepository) Create(session *domain.Session, ctx context.Context) (int64, error) {
	conn, err := ses.dbPool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin create new session tx: %w", err)
	}
	var sessionId int64

	err = conn.QueryRow(ctx, // RETURNING id: This tells PostgresSQL to return the value of the id column after insertion
		`
INSERT INTO  sessions (name, user_fp, asset_id, created_at, end_time, start_time, status,
                       current_highest_bid, auction_type, reserve_price, auto_execute, bid_increment_amount)
VALUES ($1, $2, $3,  $4, $5, $6, $7, $8, $9, $10, $11,  $12)
RETURNING id
`,
		session.Name,
		session.UserFp,
		session.AssetId,
		session.CreatedAt,
		session.EndTime,
		session.StartTime,
		session.Status,
		session.CurrentHighestBid,
		session.ActionType,
		session.ReservePrice,
		session.AutoExecute,
		session.BidIncrementAmount,
	).Scan(&sessionId)
	if err != nil {
		if err := conn.Rollback(ctx); err != nil {
			return 0, fmt.Errorf("failed to rollback create new session tx: %w", err)
		}
		return 0, err
	}

	err = conn.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return sessionId, nil
}

func (ses *sessionRepository) FindById(sessionId string, ctx context.Context) (*domain.Session, error) {
	panic("implement me")
}

func (ses *sessionRepository) FindByAssetId(assetId string, ctx context.Context) (*domain.Session, error) {
	panic("implement me")
}

func NewSessionRepository(log slog.Logger, dbPool *pgx.Conn) SessionRepository {
	return &sessionRepository{log: log, dbPool: dbPool}
}
