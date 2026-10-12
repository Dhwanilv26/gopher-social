package main

import "net/http"

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("health check ok!"))

	app.store.Posts.Create(r.Context())
}
