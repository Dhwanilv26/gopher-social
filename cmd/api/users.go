package main

import (
	"net/http"
	"strconv"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) GetUserHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil || userID < 1 {
		app.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := app.getUser(ctx, userID)

	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
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

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) FollowUserHandler(w http.ResponseWriter, r *http.Request) {

	followerUser := getUserFromContext(r)

	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followerUser.ID, followedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.ResourceConflictResponse(w, r, err)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}

	if err := app.JSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternalServerError(w, r, err)
	}

}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	followerUser := getUserFromContext(r)

	unFollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, followerUser.ID, unFollowedID); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternalServerError(w, r, err)
	}

}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, ""); err != nil {
		app.InternalServerError(w, r, err)
	}
}
