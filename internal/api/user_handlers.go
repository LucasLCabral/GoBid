package api

import (
	"errors"
	"net/http"

	"github.com/LucasLCabral/go-bid/internal/jsonutils"
	"github.com/LucasLCabral/go-bid/internal/services"
	"github.com/LucasLCabral/go-bid/internal/usecase/user"
)

func (a *API) HandleSignUpUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":    "invalid request",
			"problems": problems,
		})
		return
	}
	id, err := a.UserService.CreateUser(r.Context(), data.UserName, data.Email, data.Password, data.Bio)
	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmailOrUserName) {
			_ = jsonutils.EncodeJson(w, r, http.StatusConflict, map[string]any{
				"error": "username or email already in use",
			})
			return
		}
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"user_id": id,
	})
}

func (a *API) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.LoginUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, problems)
	}
	id, err := a.UserService.AuthenticateUser(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"error": "invalid credentials",
			})
			return
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	err = a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	a.Sessions.Put(r.Context(), "AuthenticatedUserId", id)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "logged in successfully",
	})
}

func (a *API) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	err := a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	a.Sessions.Remove(r.Context(), "AuthenticatedUserId")

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "logged out successfully",
	})
}
