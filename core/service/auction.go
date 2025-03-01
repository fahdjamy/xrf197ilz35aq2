package service

import (
	"context"
	"xrf197ilz35aq2/internal/exchange"
)

type Auction interface {
	NewAuction(request exchange.NewAuctionRequest, ctx context.Context) error
}

type auctionService struct {
}

func (a *auctionService) NewAuction(request exchange.NewAuctionRequest, ctx context.Context) error {
	return nil
}

func NewAuctionService() Auction {
	return &auctionService{}
}
