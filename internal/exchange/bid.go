package exchange

import "time"

type BidRequest struct {
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	UserFp    string    `json:"placedBy"`
	Placed    time.Time `json:"placedAt"`
	AssetId   string    `json:"assetId"`
	LastUntil time.Time `json:"lastUntil"`
}
