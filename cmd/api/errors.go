package main

import (
	"net/http"
)

func (app *application) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err)
	WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("resource not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	WriteJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *application) ResourceConflictResponse(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	app.logger.Warnf("resource conflict", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	WriteJSONError(
		w,
		http.StatusConflict,
		"resource conflict error",
	)
}
func (app *application) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	WriteJSONError(w, http.StatusUnauthorized, "unauthorized error")
}

func (app *application) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) ForbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("forbidden", "method", r.Method, "path", r.URL.Path, "error", "forbidden")
	WriteJSONError(w, http.StatusForbidden, "forbidden")
}
