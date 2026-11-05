package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func (app *application) AuthorizePostAccess(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)

		if post.UserID == user.ID {
			// returning because, we (users) can do CRUD on OUR posts
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.InternalServerError(w, r, err)
			return
		}

		if !allowed {
			app.ForbiddenResponse(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {

	role, err := app.store.Roles.GetByName(ctx, roleName)

	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.UnauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			app.UnauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.UnauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.UnauthorizedErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.getUser(ctx, userID)
		// storing only the userID in the jwt sub, but keeping the whole user in the authcontext, by getting the user from the userID
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) BasicAuthMiddlware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.UnauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.UnauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.UnauthorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			password := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)

			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.UnauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {

	// If Redis caching is disabled, directly query PostgreSQL.
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetById(ctx, userID)
	}

	// Try retrieving the user from Redis.
	user, err := app.cacheStorage.Users.Get(ctx, userID)

	if err == nil {
		// Cache HIT
		app.logger.Infow("cache hit", "key", "user", "id", userID)
		return user, nil
	}

	// Cache MISS or Redis error
	if errors.Is(err, redis.Nil) {
		app.logger.Infow("cache miss", "key", "user", "id", userID)
	} else {
		app.logger.Errorw("redis cache error", "id", userID, "error", err)
	}

	// Fetch from PostgreSQL.
	user, err = app.store.Users.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate Redis cache.
	if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
		app.logger.Errorw("failed to cache user", "id", userID, "error", err)
	}

	return user, nil

}
