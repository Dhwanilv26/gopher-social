package main

import "net/http"

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// RegisterUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	store.User			"User registered"
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := ReadJSON(w, r, payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
}
