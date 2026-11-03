package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dhwanilv26/gopher-social/docs" // this is required to generate swagger docs
	"github.com/Dhwanilv26/gopher-social/internal/auth"
	"github.com/Dhwanilv26/gopher-social/internal/mailer"
	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	env         string
	db          dBConfig
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
}
type authConfig struct {
	basic basicConfig
	token tokenConfig
}
type tokenConfig struct {
	secret string
	exp    time.Duration
	issuer string
}
type basicConfig struct {
	user string
	pass string
}

type mailConfig struct {
	sendGrid  sendGridConfig
	exp       time.Duration
	fromEmail string
}

type sendGridConfig struct {
	apiKey string
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

		r.With(app.BasicAuthMiddlware()).Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.CreatePostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.PostsContextMiddleware)
				r.Get("/", app.GetPostHandler)
				r.Delete("/", app.DeletePostHandler)
				r.Patch("/", app.UpdatePostHandler)
			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/", app.GetUserHandler)
				r.Put("/follow", app.FollowUserHandler)
				r.Put("/unfollow", app.UnfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.GetUserFeedHandler)
			})
		})

		// public routes

		r.Route("/authentication", func(r chi.Router) {
			// registration endpoint
			r.Post("/user", app.RegisterUserHandler)
			// login endpoint
			r.Post("/token", app.CreateTokenHandler)
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
