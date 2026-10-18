package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}
	if err := WriteJSON(w, http.StatusOK, data); err != nil {
		WriteJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
