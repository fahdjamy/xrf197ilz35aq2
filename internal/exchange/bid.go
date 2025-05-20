package exchange

import "time"

type BidRequest struct {
	Amount    float64   `json:"amount"`
	UserFp    string    `json:"placedBy"`
	AssetId   string    `json:"assetId"`
	LastUntil time.Time `json:"lastUntil"`
}
