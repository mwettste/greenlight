package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwettste/greenlight/internal/data"
)

func (app *application) createMovieHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "create new movie")
}

func (app *application) showMovieHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := app.readIDParameter(request)
	if err != nil {
		http.NotFound(writer, request)
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
		app.logger.Println(err)
		http.Error(writer, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
	}
}
