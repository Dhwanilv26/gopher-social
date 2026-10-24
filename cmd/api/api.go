package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr string
	env  string
	db   dBConfig
}

type dBConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.CreatePostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.PostsContextMiddleware)
				r.Get("/", app.GetPostHandler)
				r.Delete("/", app.DeletePostHandler)
				r.Patch("/", app.UpdatePostHandler)
			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.UserContextMiddleware)

				r.Get("/", app.GetUserHandler)
				r.Put("/follow", app.FollowUserHandler)
				r.Put("/unfollow", app.UnfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.GetUserFeedHandler)
			})
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("server has started at %s", app.config.addr)

	return srv.ListenAndServe()
}
