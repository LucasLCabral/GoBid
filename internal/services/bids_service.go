package services

import (
	"context"
	"errors"

	"github.com/LucasLCabral/go-bid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BidsService struct {
	queries *pgstore.Queries
	pool    *pgxpool.Pool
}

func NewBidsService(pool *pgxpool.Pool) *BidsService {
	return &BidsService{
		queries: pgstore.New(pool),
		pool:    pool,
	}
}

var ErrBidAmountTooLow = errors.New("bid amount must be greater than the base price and the highest bid")

func (bs *BidsService) PlaceBid(ctx context.Context, product_id, bidder_id uuid.UUID, bid_amount float64) (pgstore.Bid, error) {
	// ammount > previus_amount
	// ammount > baseprice
	product, err := bs.queries.GetProductByID(ctx, product_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	highestBid, err := bs.queries.GetHighestBidByProductID(ctx, product_id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.BasePrice >= bid_amount || highestBid.BidAmount >= bid_amount {
		return pgstore.Bid{}, ErrBidAmountTooLow
	}
	highestBid, err = bs.queries.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID: product_id,
		BidderID: bidder_id,
		BidAmount: bid_amount,
	})
	if err != nil {
		return pgstore.Bid{}, err
	}
	return highestBid, nil
}
