package domain

import "time"

// Session captures the session for which bids can be placed on an asset. Think of it as an auction span.
// e.g a trading day or session could be considered a bidding session.
// Bids and asks (offers) are placed and matched within a trading session.
// this table provides a grouping for bids that belong to a specific bidding event for an asset.
type Session struct {
	Name              string    `json:"name"`
	AssetId           string    `json:"assetId"`
	UserFp            string    `json:"userFp"`
	SessionId         string    `json:"sessionId"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	EndTime           time.Time `json:"endTime"`
	StartTime         time.Time `json:"startTime"`
	Status            string    `json:"status"` // ["Scheduled," "Active," "Closed," "Completed," "Cancelled."]
	CurrentHighestBid string    `json:"currentHighestBid"`
	// defines the format/rules of the auction. Different auction types have different bidding mechanisms & strategies.
	ActionType string `json:"actionType"`
	// Allows asset owners to set a minimum value they are willing to accept
	ReservePrice float64 `json:"reservePrice"` // The minimum price the product must reach for a sale to occur.
	AutoExecute  bool    `json:"autoExecute"`  // seal asset if true and contract holds plus bis rules.
	// The bidIncrementAmount is the min amount by w/c a new bid must exceed the currentHighestBid. For EnglishAuction/ascending auctions
	BidIncrementAmount float64 `json:"bidIncrementAmount"`
}

func IsValidAuctionType(auctionType string) bool {
	if auctionType == "" {
		return false
	}
	auctionTypes := make([]string, 0)
	auctionTypes = append(auctionTypes, EnglishAuction, DutchAuction, SealedAuction, FirstPriceSealedAuction, FixedPriceAuction)
	for _, aucType := range auctionTypes {
		if aucType == auctionType {
			return true
		}
	}
	return false
}
