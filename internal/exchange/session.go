package exchange

import "time"

type NewSessionRequest struct {
	AssetId            string    `json:"assetId" validate:"required"`
	Name               string    `json:"name"`
	EndTime            time.Time `json:"endTime"  validate:"required"`
	StartTime          time.Time `json:"startTime"  validate:"required"`
	ReservePrice       float64   `json:"reservePrice"`
	AutoExecute        bool      `json:"autoExecute"`
	BidIncrementAmount float64   `json:"bidIncrementAmount" validate:"required,numeric,gt=0"`
	Type               string    `json:"type"  validate:"auctionType"`
}
