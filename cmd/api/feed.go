package main

import "net/http"

func (app *application) GetUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(42))

	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, feed); err != nil {
		app.InternalServerError(w, r, err)
	}
}
