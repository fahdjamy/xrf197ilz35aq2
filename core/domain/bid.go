package domain

import (
	"fmt"
	"time"
)

const (
	PendingBid  = "PENDING"
	RejectedBid = "REJECTED"
	AcceptedBid = "ACCEPTED"
)

type Bid struct {
	Id        int64     `json:"bidId" db:"id"`
	Amount    float64   `json:"amount" db:"amount"`
	AssetId   string    `json:"assetId" db:"asset_id"`
	Status    string    `json:"status" db:"bid_status"`
	Accepted  bool      `json:"accepted" db:"accepted"`
	UserFp    string    `json:"placedBy" db:"placed_by"`
	Placed    time.Time `json:"placedAt" db:"placed_at"`
	LastUntil time.Time `json:"lastUntil" db:"last_until"`
	SessionId int64     `json:"sessionId" db:"session_id"`
}

func NewBid(userFp string, amount float64, assetId string, lastUntil time.Time, sessionId int64) (*Bid, error) {
	now := time.Now()
	if isNotValidLastingTime(lastUntil) {
		return nil, fmt.Errorf("lasting time %s is not a valid lasting time", lastUntil)
	}
	return &Bid{
		Placed:    now,
		Amount:    amount,
		Accepted:  false,
		UserFp:    userFp,
		AssetId:   assetId,
		SessionId: sessionId,
		LastUntil: lastUntil,
		Status:    PendingBid,
	}, nil
}

func isNotValidLastingTime(lastUntil time.Time) bool {
	now := time.Now()
	return lastUntil.Before(now)
}

func IsValidBidStatus(status string) bool {
	if status == "" {
		return false
	}
	bidStatuses := make([]string, 0)
	bidStatuses = append(bidStatuses, PendingBid, RejectedBid, AcceptedBid)

	for _, bStatus := range bidStatuses {
		if bStatus == status {
			return true
		}
	}
	return false
}
