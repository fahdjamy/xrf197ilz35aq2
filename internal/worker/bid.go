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
	"xrf197ilz35aq2/storage/timescale/queries"
)

type JobsConfig struct {
	timeout time.Duration
	sleep   time.Duration
}

type BidWorker struct {
	log       slog.Logger
	client    *redis.Client
	timeout   time.Duration
	sleep     time.Duration
	tsQuerier queries.BidTSQuerier
	bidRepo   postgres.BidRepository
}

// ProcessCachedBidsFromQueue --- Background Worker (separate process or goroutine)
// Start this in a `main` function, likely as a goroutine.
func (worker *BidWorker) ProcessCachedBidsFromQueue(ctx context.Context, queue string) error {
	for {
		// 1. Fetch bids from a cache
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

		// 2. Process cached string bids and unmarshal them into Bid struct
		bids := make([]domain.Bid, 0)
		for _, cachedBid := range result {
			var bid domain.Bid
			if err := json.Unmarshal([]byte(cachedBid), &bid); err != nil {
				worker.log.Error("Error unmarshalling bid from queue: %v", "err", err)
				break
			}
			bids = append(bids, bid)
		}

		// only start saving bids if at least more than one bid is unmarshalled.
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
		worker.logSavedBids("postgres", err, count, int64(len(bids)))

		// 4. Save bids to the timescale db as well
		tsRespCnt, err := worker.tsQuerier.BatchSaveBid(ctx, bids)
		if err != nil {
			worker.log.Error("Error saving bids to timescale", "err", err)
			time.Sleep(worker.sleep)
			continue
		}

		// 5. Log successfully stored bids
		worker.logSavedBids("timescale", err, tsRespCnt, int64(len(bids)))
		time.Sleep(worker.sleep)
		continue
	}
}

func (worker *BidWorker) logSavedBids(dbTye string, err error, count int64, expectedCnt int64, args ...interface{}) {
	if err != nil {
		worker.log.Error(fmt.Sprintf("Error saving bids to %s", dbTye), "err", err)
	} else if count != expectedCnt {
		worker.log.Warn(fmt.Sprintf("Error saving bids to %s: savedBidsCount=%d of bids cachedCount=%d", dbTye, count, expectedCnt), args...)
	} else {
		worker.log.Info(fmt.Sprintf(fmt.Sprintf("Successfully saved %d bids to %s", count, dbTye), args...))
	}
}

func NewBidWorker(log slog.Logger, client *redis.Client, config JobsConfig, bidRepo postgres.BidRepository, tsQuerier queries.BidTSQuerier) *BidWorker {
	return &BidWorker{
		log:       log,
		client:    client,
		bidRepo:   bidRepo,
		tsQuerier: tsQuerier,
		sleep:     config.sleep,
		timeout:   config.timeout,
	}
}
