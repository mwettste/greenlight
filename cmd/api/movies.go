package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwettste/greenlight/internal/data"
)

func (app *application) createMovieHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	fmt.Fprintf(writer, "%+v\n", input)
}

func (app *application) showMovieHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := app.readIDParameter(request)
	if err != nil {
		app.notFoundResponse(writer, request)
		return
	}

	movie := data.Movie{
		ID:         id,
		CreatedAt:  time.Now(),
		Title:      "Lord of the Rings - The Fellowship of the Ring",
		RuntimeMin: 178,
		Genres:     []string{"fantasy", "adventure"},
		Version:    1,
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}
