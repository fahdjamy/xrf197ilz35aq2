package domain

import "time"

const PendingBid = "PENDING"

type Bid struct {
	UserFp    string    `json:"userFp"`
	Price     float64   `json:"price"`
	Placed    time.Time `json:"placedAt"`
	AssetId   string    `json:"assetId"`
	Accepted  bool      `json:"accepted"`
	Status    string    `json:"status"`
	SessionId string    `json:"sessionId"`
}

func NewBid(userFp string, price float64, assetId string) *Bid {
	now := time.Now()
	return &Bid{
		Placed:   now,
		UserFp:   userFp,
		Price:    price,
		AssetId:  assetId,
		Accepted: false,
		Status:   "",
	}
}
