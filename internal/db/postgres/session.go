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
	FindAllByAssetId(assetId string, ctx context.Context) ([]domain.Session, error)
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
INSERT INTO  sessions (session_name, user_fp, asset_id, created_at, end_time, start_time, status,
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
	session := &domain.Session{}
	err := ses.dbPool.QueryRow(ctx, `
SELECT id, 
       auto_execute, 
       user_fp, 
       asset_id,
       status, 
       session_name, 
       reserve_price,
       auction_type,
       end_time,
       start_time,
       created_at,
       current_highest_bid,
       bid_increment_amount
FROM sessions
WHERE id = $1`, sessionId).Scan(
		&session.Id,
		&session.AutoExecute,
		&session.UserFp,
		&session.AssetId,
		&session.Status,
		&session.Name,
		&session.ReservePrice,
		&session.ActionType,
		&session.EndTime,
		&session.StartTime,
		&session.CreatedAt,
		&session.CurrentHighestBid,
		&session.BidIncrementAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find session by id: %w", err)
	}

	return session, nil
}

func (ses *sessionRepository) FindAllByAssetId(assetId string, ctx context.Context) ([]domain.Session, error) {
	sql := `
SELECT 
	id, auto_execute, user_fp, asset_id, status, session_name, reserve_price, auction_type, end_time, created_at, 
	current_highest_bid, bid_increment_amount
FROM sessions
WHERE asset_id = $1
`
	rows, err := ses.dbPool.Query(ctx, sql, assetId)
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions by asset id: %w", err)
	}
	var sessions []domain.Session
	for rows.Next() {
		var session domain.Session
		err = rows.Scan(
			&session.Id,
			&session.AutoExecute,
			&session.UserFp,
			&session.AssetId,
			&session.Status,
			&session.Name,
			&session.ReservePrice,
			&session.ActionType,
			&session.EndTime,
			&session.CreatedAt,
			&session.CurrentHighestBid,
			&session.BidIncrementAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning of the sessions: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func NewSessionRepository(log slog.Logger, dbPool *pgx.Conn) SessionRepository {
	return &sessionRepository{log: log, dbPool: dbPool}
}
