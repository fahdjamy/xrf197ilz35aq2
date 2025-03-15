package exchange

import "time"

type BidRequest struct {
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	UserFp    string    `json:"placedBy"`
	Placed    time.Time `json:"placedAt"`
	SessionId int64     `json:"sessionId"`
	LastUntil time.Time `json:"lastUntil"`
}
