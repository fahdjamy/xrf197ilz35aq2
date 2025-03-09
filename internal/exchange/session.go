package exchange

import "time"

type NewSessionRequest struct {
	AssetId            string    `json:"assetId" validate:"required"`
	EndTime            time.Time `json:"endTime"  validate:"required"`
	StartTime          time.Time `json:"startTime"  validate:"required"`
	ReservePrice       int       `json:"reservePrice"`
	AutoExecute        bool      `json:"autoExecute"`
	BidIncrementAmount float64   `json:"bidIncrementAmount" validate:"required,numeric,gt=0"`
	Type               string    `json:"type"  validate:"auctionType"`
}
