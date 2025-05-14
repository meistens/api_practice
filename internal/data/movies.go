package data

import (
	"context"
	"database/sql"
	"errors"
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

	// create context with 3s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// queryrowcontext() to execute the sql on conn. pool, pass the ctx as first arg
	// passing int the args slice as a varidic param and scanning
	// the generated id, created_at and version into the movie struct
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// placeholder for fetchig specific record
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// define sql query for retrieving data
	query := `SELECT id, created_at, title, year, runtime, genres, version FROM movies WHERE id = $1`

	// declare movie struct to hold the data returned by query
	var movie Movie

	// query timeout using context.withtimeout() func. to create a timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer to cancel context before Get() returns
	defer cancel()

	// use queryrowcontext() to execute query, passing in
	// the contect with the deadline as dirst arg
	// scan the response data into the field of the movie struct
	// use pq.array() to convert the scan target for genres column
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)

	// err handling, if no matching movie found, scan will return
	// a sql.errnorows
	// instead we use the custom error instead
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// otherwise return a pointer to the movie struct
	return &movie, nil
}

// placeeholder for updating a specific record
func (m MovieModel) Update(movie *Movie) error {
	// TODO: implement uuid for version
	query := `UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND VERSION =$6
	RETURNING version`

	// create an args slice containing the values for the placeholder params
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// context...
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// queryrowcontext()
	// execute sql query, if no matching rows found, then movie version
	// has changed (or has been deleted), return to errConflict
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// placeholder for deleting a specific record
func (m MovieModel) Delete(id int64) error {
	// return error if id less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM movies
	WHERE id = $1`

	// context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// execute query using ExecContext() method, passing the id variable as
	// the value for the placeholder param
	// it returns an sql.result object
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// call rowsaffected() on the sql.result object
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// if no rows affected, it means no records in the movie table
	// so error
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// GetAll func, returns a slice of movies
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// filter conditions for title and genres, can be used singly or at once
	query := `SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY id`

	// create a context with 3s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// querycontext() to exec. query, returns a sql.rows resultset
	// pass the title and genres as placeholder params or you'd be getting
	// missing parameters from postgres
	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}
	// defer call to rows.close() to ensure resultset closed before getall() returns
	defer rows.Close()

	// init. empty slice to hold movie data
	movies := []*Movie{}

	// use rows.Next to iterate through rows in resultset
	for rows.Next() {
		// init. empty movie struct
		var movie Movie

		// scan values from row into Movie struct
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		// add Movie struct to slice
		movies = append(movies, &movie)
	}
	// when rows.next() loop is done, call rows.err() to get any error
	// thrown during its iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// slice should be returned if everything ok
	return movies, nil
}
