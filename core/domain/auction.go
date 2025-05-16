package domain

// **EnglishAuction** (or Ascending). Bidders openly bid against each other, with each bid needing than the previous one.
// Auction ends at a predefined time, and the highest bidder wins.
// **DutchAuction** The auction starts at a high price, which is gradually lowered until a bidder accepts the current price.
// The first bidder to accept wins and pays that price for each item in the lot.
// **SealedAuction** Bidders submit their bids privately, without knowing others' bids. At the end of the bidding
// period, the bids are revealed, and the highest bidder wins
// **FirstPriceSealedAuction**: Highest bidder wins and pays their bid.
// **FixedPriceAuction** Asset is sold at a set price, no bidding involved.
const (
	DutchAuction            = "DutchAuction"
	SealedAuction           = "SealedAuction"
	EnglishAuction          = "EnglishAuction" // The new bid must be higher than the current highest bid bid_increment_amount
	FixedPriceAuction       = "FixedPriceAuction"
	FirstPriceSealedAuction = "FirstPriceSealedAuction"
)
