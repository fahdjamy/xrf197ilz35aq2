CREATE TABLE IF NOT EXISTS bid_records (
    id VARCHAR(255) NOT NULL,
    symbol TEXT NOT NULL,
    is_accepted BOOLEAN,
    asset_id VARCHAR(255) NOT NULL,
    bidder_fp VARCHAR(255) NOT NULL,
    seller_fp VARCHAR(255) NOT NULL,
    trade_time TIMESTAMPTZ NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    expiration_time TIMESTAMPTZ NOT NULL
);
