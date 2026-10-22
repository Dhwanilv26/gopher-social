package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()

	user, err := app.store.Users.GetById(ctx, userId)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.BadRequestResponse(w, r, err)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}

	if err := app.JSONResponse(w, http.StatusOK, user); err != nil {
		app.InternalServerError(w, r, err)
	}

}
