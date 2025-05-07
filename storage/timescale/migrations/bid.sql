CREATE TABLE IF NOT EXISTS bid_records (
    accepted BOOLEAN,
    symbol TEXT accepted,
    asset_id VARCHAR(255) NOT NULL,
    trade_time TIMESTAMPTZ NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    buyer_user_fp VARCHAR(255) NOT NULL,
    buyer_user_fp VARCHAR(255) NOT NULL,
    expiration_time TIMESTAMPTZ NOT NULL
);
