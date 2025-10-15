package product

import (
	"context"
	"time"

	"github.com/LucasLCabral/go-bid/internal/validator"
	"github.com/google/uuid"
)

type CreateProductReq struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionEnd = 2 * time.Hour

func (req CreateProductReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "must be provided")
	eval.CheckField(validator.NotBlank(req.Description), "description", "must be provided")
	eval.CheckField(
		validator.MinChars(req.Description, 10) &&
			validator.MaxChars(req.Description, 255),
		"description", "must be between 10 and 255 characters long",
	)
	eval.CheckField(req.BasePrice > 0, "base_price", "must be greater than 0")
	eval.CheckField(time.Until(req.AuctionEnd) >= minAuctionEnd, "auction_end", "must be at least 2 hours from now")
	return eval
}
