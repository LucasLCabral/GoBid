package api

import (
	"net/http"

	"github.com/LucasLCabral/go-bid/internal/jsonutils"
	"github.com/LucasLCabral/go-bid/internal/usecase/product"
	"github.com/google/uuid"
)

func (a *API) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, problems)
		return
	}
	userID, ok := a.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": "must be logged in",
		})
		return
	}
	id, err := a.ProductsService.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd,
	)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}
	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "product created successfully",
		"product_id": id,
	})
}
