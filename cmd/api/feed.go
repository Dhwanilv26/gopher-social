package main

import "net/http"

func (app *application) GetUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	type feedUserPayload struct {
		userID int64
	}

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(26))

	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, feed); err != nil {
		app.InternalServerError(w, r, err)
	}
}
