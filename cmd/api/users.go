package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mwettste/greenlight/internal/data"
	"github.com/mwettste/greenlight/internal/validator"
)

func (app *application) registerUserHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(writer, request, &input)

	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(writer, request, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(writer, request, v.Errors)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"userID":          user.ID,
			"userName":        user.Name,
			"activationToken": token.Plaintext,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
		app.logger.PrintInfo("successfully sent registration e-mail", map[string]string{"email": user.Email})
	})

	err = app.writeJSON(writer, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}

}

func (app *application) activateUserHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(writer, request, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(writer, request, v.Errors)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(writer, request)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}

func (app *application) updateUserPasswordHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Password       string `json:"password"`
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlaintext(v, input.TokenPlaintext)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(writer, request, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(writer, request, v.Errors)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(writer, request)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	env := envelope{"message": "your password was successfully reset"}

	err = app.writeJSON(writer, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}
