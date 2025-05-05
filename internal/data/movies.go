package data

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/meistens/api_practice/internal/validator"
)

// NOTE -> CAPITALIZE FIRST LETTER MEANS TO BE VISIBLE WHEN EXPORTED!
type Movie struct {
	ID        int64     `json:"id"`             // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`              // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`          // Movie title
	Year      int32     `json:"year,omitempty"` // Movie release year
	// use Runtime type here, doing this is to wrap the int in a double quote
	// while still making it an int (don't think too much about this... remove if it will be uncomfortable to use)
	Runtime Runtime  `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres  []string `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc.)
	Version int32    `json:"version"`           // The version number starts at 1 and will be incremented each
	// time the movie information is updated
}

// struct wraps a conn. pool
type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// accepts a pointer to a movie struct, which creates data for the new record
func (m MovieModel) Insert(movie *Movie) error {
	// define sql query for inserting a new record in the movies table
	// returns system-generated data
	query := `INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	// create an arg slice containing the values for the placeholder params
	// from the movie struct
	// Declaring the slice immediately next to sql query helps
	// make it nice and clear **what values are being used where**
	// in the query
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// queryrow() to execute the sql on conn. pool
	// passing int the args slice as a varidic param and scanning
	// the generated id, created_at and version into the movie struct
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// placeholder for fetchig specific record
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// placeeholder for updating a specific record
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// placeholder for deleting a specific record
func (m MovieModel) Delete(id int64) error {
	return nil
}
