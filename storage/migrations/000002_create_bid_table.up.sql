-- Create the ENUM for BidStatus
CREATE TYPE bid_status AS ENUM (
    'PENDING',
    'REJECTED',
    'ACCEPTED'
);

CREATE TABLE IF NOT EXISTS asset_bid (
    id SERIAL PRIMARY KEY,
    accepted BOOLEAN NOT NULL,
    status bid_status NOT NULL,
    asset_id VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    placed_by VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    last_until TIMESTAMP WITH TIME ZONE NOT NULL,
    placed_at TIMESTAMP WITH TIME ZONE NOT NULL
);
