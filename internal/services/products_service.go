package services

import (
	"context"
	"time"

	"github.com/LucasLCabral/go-bid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductsService(pool *pgxpool.Pool) *ProductsService {
	return &ProductsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductsService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName,
	description string,
	baseprice float64,
	auctionEnd time.Time,
) (uuid.UUID, error) {
	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: productName,
		Description: description,
		BasePrice:   baseprice,
		AuctionEnd:  auctionEnd,
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}
