package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
} // only dealing with the fields client is allowed to send while creating a post

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
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

	if err := app.JSONResponse(w, http.StatusCreated, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

func (app *application) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID)

	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

func (app *application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postId")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r)

	var payload UpdatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
		app.InternalServerError(w, r, err)
	}
}

func (app *application) PostsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idParam := chi.URLParam(r, "postId")
		id, err := strconv.ParseInt(idParam, 10, 64)

		if err != nil {
			app.BadRequestResponse(w, r, err)
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
		// creating new context with key "post" and value -> post struct
		ctx = context.WithValue(ctx, postCtx, post)
		// telling the middlware to proceed with the next handler, with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
