package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
	"xrf197ilz35aq2/core/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (string, error)
	FindById(ctx context.Context, sessionId string) (*domain.Session, error)
	FindActiveSession(ctx context.Context, assetId string) (*domain.Session, error)
	FindAllByAssetId(ctx context.Context, assetId string) ([]domain.Session, error)
}

type sessionRepository struct {
	log    slog.Logger
	dbPool *pgxpool.Pool
}

func (ses *sessionRepository) Create(ctx context.Context, session *domain.Session) (string, error) {
	conn, err := ses.dbPool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin create new session tx: %w", err)
	}
	var sessionId string

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
			return "", fmt.Errorf("failed to rollback create new session tx: %w", err)
		}
		return "", err
	}

	err = conn.Commit(ctx)
	if err != nil {
		return "", err
	}
	return sessionId, nil
}

func (ses *sessionRepository) FindById(ctx context.Context, sessionId string) (*domain.Session, error) {
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

func (ses *sessionRepository) FindAllByAssetId(ctx context.Context, assetId string) ([]domain.Session, error) {
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
	defer rows.Close()
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

func (ses *sessionRepository) FindActiveSession(ctx context.Context, assetId string) (*domain.Session, error) {
	now := time.Now()
	sql := `
SELECT id, auto_execute, user_fp, asset_id, status, session_name, reserve_price, auction_type, end_time, created_at, 
	current_highest_bid, bid_increment_amount
FROM sessions
WHERE asset_id = $1
AND end_time > $2`

	rows, err := ses.dbPool.Query(ctx, sql, assetId, now)
	if err != nil {
		return nil, fmt.Errorf("failed to find active session: %w", err)
	}
	defer rows.Close()
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
	if len(sessions) == 0 {
		return nil, fmt.Errorf("there are no active sessions for the asset")
	}
	if len(sessions) > 1 {
		return nil, fmt.Errorf("invalid session state, found more than one active sessions for the asset")
	}
	return &sessions[0], nil
}

func NewSessionRepository(dbPool *pgxpool.Pool, log slog.Logger) SessionRepository {
	return &sessionRepository{log: log, dbPool: dbPool}
}
