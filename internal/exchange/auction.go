package exchange

import "time"

type NewAuctionRequest struct {
	AssetId            string    `json:"assetId"`
	EndTime            time.Time `json:"endTime"`
	StartTime          time.Time `json:"startTime"`
	Type               string    `json:"type"`
	ReservePrice       int       `json:"reservePrice"`
	AutoExecute        bool      `json:"autoExecute"`
	BidIncrementAmount float64   `json:"bidIncrementAmount"`
}
