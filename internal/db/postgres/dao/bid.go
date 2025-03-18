package dao

// BidColumns makes sure they match the column names defined in the migration file for bid
const BidColumns = `
id, accepted, status, asset_id, amount, placed_by, session_id, last_until, placed_at
`
