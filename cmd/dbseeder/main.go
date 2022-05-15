package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mwettste/greenlight/internal/data"
)

var sampleMovies = []data.Movie{
	{
		Title:      "Moana",
		Year:       2016,
		RuntimeMin: data.RuntimeMin(107),
		Genres:     []string{"animation", "adventure"},
	},
	{
		Title:      "Black Panther",
		Year:       2018,
		RuntimeMin: data.RuntimeMin(134),
		Genres:     []string{"action", "adventure"},
	},
	{
		Title:      "Deadpool",
		Year:       2016,
		RuntimeMin: data.RuntimeMin(108),
		Genres:     []string{"action", "comedy"},
	},
	{
		Title:      "The Breakfast Club",
		Year:       1986,
		RuntimeMin: data.RuntimeMin(107),
		Genres:     []string{"drama"},
	},
}

var filters = data.Filters{
	Page:         1,
	PageSize:     5,
	Sort:         "id",
	SortSafelist: []string{"id"},
}

func main() {
	var dsn string
	flag.StringVar(&dsn, "dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.Parse()
	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	models := data.NewModels(db)

	noOfMovies := len(sampleMovies)
	for i, movie := range sampleMovies {
		fmt.Printf("Inserting movie %d of %d with title %s\n", i+1, noOfMovies, movie.Title)
		_, metadata, err := models.Movies.GetAll(movie.Title, []string{}, filters)

		if err != nil {
			log.Fatalf("Failed to check if movie exists: %v\n", err)
		}

		if metadata.TotalRecords > 1 {
			log.Fatalf("Found multiple records matching the title: %s\n", movie.Title)
		}

		if metadata.TotalRecords == 1 {
			fmt.Println("Movie already in DB...")
			continue
		}

		err = models.Movies.Insert(&movie)
		if err != nil {
			log.Fatalf("Failed to insert movie: %v\n", err)
		}

		fmt.Println("Successfully inserted movie...")
	}

	fmt.Println("All done - your DB is now seeded with some sample movies!")
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
