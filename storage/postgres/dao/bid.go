package dao

// BidTableName makes sure they match the column names defined in the migration file for bid
const BidTableName = "asset_bid"

func GetBidColumnName() []string {
	return []string{
		"id",
		"accepted",
		"status",
		"asset_id",
		"amount",
		"placed_by",
		"session_id",
		"last_until",
		"placed_at",
	}
}
