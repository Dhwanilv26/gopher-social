package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/Dhwanilv26/gopher-social/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()

	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()

	return &application{
		logger:       logger,
		store:        mockStore,
		cacheStorage: mockCacheStore,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {

	fmt.Println("REQUEST PATH:", req.URL.Path)
	fmt.Println("REQUEST METHOD:", req.Method)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}
