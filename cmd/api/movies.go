package main

import (
	"fmt"
	"net/http"
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
	fmt.Fprintf(writer, "show details of the movie %d\n", id)
}
