-- Create the ENUM for BidStatus
CREATE TYPE bid_status AS ENUM (
    'PENDING',
    'REJECTED',
    'ACCEPTED'
);

CREATE TABLE IF NOT EXISTS asset_bid (
    accepted BOOLEAN NOT NULL,
    status bid_status NOT NULL,
    id VARCHAR(255) PRIMARY KEY,
    asset_id VARCHAR(255) NOT NULL,
    placed_by VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    last_until TIMESTAMP WITH TIME ZONE NOT NULL,
    placed_at TIMESTAMP WITH TIME ZONE NOT NULL
);
