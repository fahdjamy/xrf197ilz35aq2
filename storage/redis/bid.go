package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/exchange"
)

type BidCache interface {
	SaveBid(request exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) (*domain.Bid, error)
}

type bidCache struct {
	log    slog.Logger
	client *redis.Client
}

func (cache *bidCache) SaveBid(request exchange.BidRequest, userFp string, sessionId int64, ctx context.Context) (*domain.Bid, error) {
	if request.Amount <= 0 {
		return nil, errors.New("invalid amount")
	}
	assetId := request.AssetId
	newBid, err := domain.NewBid(userFp, request.Amount, assetId, request.LastUntil, sessionId)
	if err != nil {
		return nil, fmt.Errorf("creating new bid failed with err=%w", err)
	}

	// 1. Push Bid to Redis Queue
	bidJSON, err := json.Marshal(newBid)
	if err != nil {
		return nil, fmt.Errorf("marshaling new bid failed with err=%w", err)
	}
	// RPush inserts all the specified values at the tail of the list stored at the key. If the key does not exist,
	// it is created as an empty list before performing the push operation. When the key holds a value that is not a list, an error is returned
	err = cache.client.RPush(ctx, bidKey(assetId, sessionId, request.LastUntil), bidJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("saving new bid failed with err=%w", err)
	}
	return newBid, nil
}

func bidKey(assetId string, sessionId int64, sessionEndTime time.Time) string {
	return fmt.Sprintf("bid_%s_%d_%d", assetId, sessionEndTime.UnixMilli(), sessionId)
}

func NewBidCache(log slog.Logger, client *redis.Client) BidCache {
	return &bidCache{
		log:    log,
		client: client,
	}
}
