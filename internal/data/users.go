package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/meistens/api_practice/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// custom errduplicateemail error
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// userModel struct which wraps the conn. pool
type UserModel struct {
	DB *sql.DB
}

var AnonUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// password type struct
// plaintext is a ptr to a string, makes distinguish
// between plaintext password not present in struct and
// plaintext in empty string
type password struct {
	plaintext *string
	hash      []byte
}

// check if a user instance is the anon
func (u *User) IsAnon() bool {
	return u == AnonUser
}

// set() calculates bcrypt hash of plaintext password
// stores both hash and plaintext in struct
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// matches() method checks if the provided plaintext matches the
// hashed password stored in the struct, returning true if it
// matches and false if it doesn't
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// insert new user record to db
func (m UserModel) Insert(user *User) error {
	query := `INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// if table already contains a record with this email, and
	// an insert is attempted, error but in a dignified manner
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {

		// Temporary debugging - remove this after fixing
		fmt.Printf("Database error: %s\n", err.Error())

		switch {
		// More robust check for duplicate email constraint violations
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "email"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// retrieve user details from db based on user email address
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// update details for a specfic user
func (m UserModel) Update(user *User) error {
	query := `UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {

		// Temporary debugging - remove this after fixing
		fmt.Printf("Database error: %s\n", err.Error())

		switch {
		// More robust check for duplicate email constraint violations
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "email"):
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// calc. hash of plaintext token provided by the client
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// setup query
	query := `
	SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1
	AND tokens.scope = $2
	AND tokens.expiry > $3
	`

	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// execute the query, scanning the return values into a User struct
	// if no matching found, errrecordnotfound()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// return the matching user
	return &user, nil
}
