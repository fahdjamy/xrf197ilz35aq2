package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/storage/postgres"
)

type BidWorker struct {
	log     slog.Logger
	client  *redis.Client
	timeout time.Duration
	sleep   time.Duration
	bidRepo postgres.BidRepository
}

// ProcessCachedBidsFromQueue --- Background Worker (separate process or goroutine)
// Start this in a `main` function, likely as a goroutine.
func (worker *BidWorker) ProcessCachedBidsFromQueue(ctx context.Context, queue string) error {
	for {
		// 1. Fetch bids from cache
		result, err := worker.client.BLPop(ctx, worker.timeout, queue).Result()
		if err != nil {
			worker.log.Warn("Error fetching bid from queue", "err", err)
			time.Sleep(worker.sleep) // Simple backoff
			continue
		}
		// no item to process, sleep and try again later
		if len(result) == 0 {
			time.Sleep(worker.sleep)
			continue
		}

		// 2. Process cached string bids and un-marshal them into Bid struct
		bids := make([]domain.Bid, 0)
		for _, cachedBid := range result {
			var bid domain.Bid
			if err := json.Unmarshal([]byte(cachedBid), &bid); err != nil {
				worker.log.Error("Error unmarshalling bid from queue: %v", "err", err)
				break
			}
			bids = append(bids, bid)
		}

		// only start saving bids if at least more than one bid is un-marshalled.
		if len(bids) == 0 {
			time.Sleep(worker.sleep)
			continue
		}

		worker.log.Info(fmt.Sprintf("Successfully fetched %d bids from queue", len(bids)))
		// 3. Store/Save bids permanently in the DB
		count, err := worker.bidRepo.CreateBidsCopyFrom(ctx, bids)
		if err != nil {
			worker.log.Error("Error creating bids", "err", err)
			time.Sleep(worker.sleep)
			continue
		}

		// 4. Log successfully stored bids
		if count != int64(len(bids)) {
			worker.log.Warn("Error creating bids: savedBidsCount=%d of bids cachedCount=%d", "count", len(bids))
		} else {
			worker.log.Info(fmt.Sprintf("Successfully saved %d bids all from cache", count))
		}
		time.Sleep(worker.sleep)
		continue
	}
}

func NewBidWorker(log slog.Logger, client *redis.Client, timeout time.Duration, workerSleep time.Duration, bidRepo postgres.BidRepository) *BidWorker {
	return &BidWorker{
		log:     log,
		client:  client,
		bidRepo: bidRepo,
		timeout: timeout,
		sleep:   workerSleep,
	}
}
