CREATE TABLE IF NOT EXISTS bid_records (
    bid_id VARCHAR(255) NOT NULL,
    symbol TEXT NOT NULL,
    is_accepted BOOLEAN,
    bid_time TIMESTAMPTZ NOT NULL,
    asset_id VARCHAR(255) NOT NULL,
    bidder_fp VARCHAR(255) NOT NULL,
    seller_fp VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    expiration_time TIMESTAMPTZ NOT NULL
);
----
SELECT create_hypertable('bid_records', 'bid_time', if_not_exists => TRUE);
----
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint
		WHERE conname = 'bid_records_bid_id_unique' AND conrelid = 'bid_records'::regclass
	) THEN
        ALTER TABLE bid_records ADD CONSTRAINT bid_records_bid_id_time_unique UNIQUE (bid_id, bid_time);
    END IF;
END;
	$$;
