package user

import (
	"context"

	"github.com/LucasLCabral/go-bid/internal/validator"
)

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "must be provided")
	eval.CheckField(validator.NotBlank(req.Email), "email", "must be provided")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "must be a valid email address")
	eval.CheckField(validator.NotBlank(req.Bio), "bio", "must be provided")
	eval.CheckField(
		validator.MinChars(req.Bio, 10) &&
			validator.MaxChars(req.Bio, 255), "bio", "must be between 10 and 255 characters long")
	eval.CheckField(validator.MinChars(req.Password, 8), "password", "must be at least 8 characters long")

	return eval
}
