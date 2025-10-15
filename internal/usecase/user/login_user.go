package user

import (
	"context"

	"github.com/LucasLCabral/go-bid/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (req LoginUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "must be a valid email address")
	eval.CheckField(validator.NotBlank(req.Password), "password", "must be provided")

	return eval
}
