package main

import (
	"net/http"

	"github.com/Dhwanilv26/gopher-social/internal/store"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
} // only dealing with the fields client is allowed to send while creating a post

func (app *application) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
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
		WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := WriteJSON(w, http.StatusCreated, post); err != nil {
		WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
