package data

import (
	"database/sql"
	"errors"
)

// custom errRecordnotfound error will return from the Get() method
// if a movie query doesn't exist
// errEditConflict for optimistic locking to prevent a race condition
var (
	ErrRecordNotFound = errors.New("record not foud")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models struct wraps the MovieModel and UserModel
type Models struct {
	Movies MovieModel
	Users  UserModel
	Tokens TokenModel
}

// Adding New() which returns a Models struct containing the
// initalized MovieModel and UserModel instance
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
