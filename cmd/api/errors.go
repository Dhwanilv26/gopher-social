package main

import (
	"log"
	"net/http"
)

func (app *application) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path : %s error : %s", r.Method, r.URL.Path, err)
	WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("resource not found error: %s path : %s error : %s", r.Method, r.URL.Path, err)
	WriteJSONError(w, http.StatusNotFound, "not found")
}

func (app *application) ResourceConflictResponse(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	log.Printf(
		"resource conflict error: %s path: %s error: %s",
		r.Method,
		r.URL.Path,
		err,
	)

	WriteJSONError(
		w,
		http.StatusConflict,
		"resource was modified; fetch the latest version and retry",
	)
}
