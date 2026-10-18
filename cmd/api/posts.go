package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
} // only dealing with the fields client is allowed to send while creating a post

func (app *application) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	userId := 1 // fake user id to avoid foreign key constraints

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// get real user id after auth
		UserID: int64(userId),
	}
	ctx := r.Context() // context passed to cancel db operations on network or db failure

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := WriteJSON(w, http.StatusCreated, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

func (app *application) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postId")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, id)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	if err := WriteJSON(w, http.StatusOK, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}
