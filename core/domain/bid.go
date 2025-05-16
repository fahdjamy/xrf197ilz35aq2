package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"
)

const (
	PendingBid  = "PENDING"
	RejectedBid = "REJECTED"
	AcceptedBid = "ACCEPTED"
)

type Bid struct {
	Id         string    `json:"bidId" db:"id"`
	AssetOwner string    `json:"_" db:"seller_fp"`
	Amount     float64   `json:"amount" db:"amount"`
	Symbol     string    `json:"symbol" db:"symbol"`
	Quantity   float64   `json:"quantity" db:"quantity"`
	AssetId    string    `json:"assetId" db:"asset_id"`
	Status     string    `json:"status" db:"bid_status"`
	Timestamp  time.Time `json:"timestamp" db:"bid_time"`
	UserFp     string    `json:"placedBy" db:"bidder_fp"`
	Accepted   bool      `json:"accepted" db:"is_accepted"`
	SessionId  string    `json:"sessionId" db:"session_id"`
	LastUntil  time.Time `json:"lastUntil" db:"expiration_time"`
}

func NewBid(userFp string, amount float64, assetId string, lastUntil time.Time, sessionId string) (*Bid, error) {
	now := time.Now()
	if isNotValidLastingTime(lastUntil) {
		return nil, fmt.Errorf("lasting time %s is not a valid lasting time", lastUntil)
	}
	return &Bid{
		Timestamp: now,
		Amount:    amount,
		Accepted:  false,
		UserFp:    userFp,
		AssetId:   assetId,
		SessionId: sessionId,
		LastUntil: lastUntil,
		Status:    PendingBid,
		Id:        strconv.FormatInt(generateId(), 10), // an ID is generated because a bid is first cached before it's saved to the DB
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

var (
	mutex     sync.Mutex
	lastValue int64
)

func generateId() int64 {
	// Used to ensure thread-safety.
	// If multiple goroutines call this function simultaneously,
	// the mutex prevents them from interfering with each other and potentially generating the same value
	mutex.Lock()
	defer mutex.Unlock()

	// Get the current timestamp in nanoseconds
	now := time.Now().UnixNano()

	// Ensure uniqueness even if called multiple times in the same nanosecond
	if now <= lastValue {
		now = lastValue + 1
	}
	lastValue = now

	// Generate a random int64 to add further uniqueness
	// The maximum value for rand.Int is adjusted to 1<<62-1 to ensure the random part is always positive.
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1<<31-1)) // Max positive int64

	// Clear the most significant 32 bits of the timestamp to ensure it's positive after the shift
	now &= 1<<32 - 1

	// Combine timestamp and random part
	// The timestamp is shifted left by 31 bits instead of 32.
	// This leaves the most significant bit free, guaranteeing the final uniqueValue is always positive.
	uniqueValue := (now << 31) | randomPart.Int64()

	return uniqueValue
}
