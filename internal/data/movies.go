package data

import "time"

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
