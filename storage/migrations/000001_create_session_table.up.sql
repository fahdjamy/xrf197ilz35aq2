-- Create the ENUM for AuctionType
CREATE TYPE session_action_type AS ENUM (
    'DutchAuction',
    'SealedAuction',
    'EnglishAuction',
    'FixedPriceAuction',
    'FirstPriceSealedAuction'
);

CREATE TYPE session_status AS ENUM (
    'Active',
    'Closed',
    'Completed',
    'Cancelled',
    'Scheduled'
);

CREATE TABLE IF NOT EXISTS bid_session (
    id                   SERIAL PRIMARY KEY,
    auto_execute         BOOLEAN          NOT NULL,
    user_fp              VARCHAR(255)     NOT NULL,
    asset_id             VARCHAR(255)     NOT NULL,
    status session_status NOT NULL,
    session_name         VARCHAR(100)     NOT NULL,
    reserve_price DOUBLE PRECISION NOT NULL,
    auction_type session_action_type NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    current_highest_bid  DOUBLE PRECISION NOT NULL,
    bid_increment_amount DOUBLE PRECISION NOT NULL
);
