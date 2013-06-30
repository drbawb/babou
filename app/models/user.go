package models

import (
	sql "database/sql"
	errors "errors"
	fmt "fmt"

	db "github.com/drbawb/babou/lib/db"
)

// `User` model for `users`
type User struct {
	Username string

	passwordHash string
	passwordSalt string
}

// Error-codes returned from some methods that could be presented to the UI.
type UserModelError int

const (
	USERNAME_TAKEN        UserModelError = 1 << iota
	USERNAME_INVALID_CHAR                = 1 << iota
)

// Creates a new user record in the database.
// The first return parameter is an error-code that represents a non-fatal
// problem that could be presented to the user.
//
// The second return parameter is a fatal error passed up from the database layer.
func NewUser(username string, passwordHash, passwordSalt []byte) (UserModelError, error) {
	var outStatus UserModelError

	fn := func(database *sql.DB) error {
		if CountUsers(username) > 0 {
			outStatus = USERNAME_TAKEN
			return nil
		}

		stmt, err := database.Prepare("INSERT INTO users(username,passwordhash,passwordsalt) VALUES($1,$2, $3)")

		if err != nil {
			return errors.New(fmt.Sprintf("Error preparing statement: %s", err.Error()))
		}

		res, err := stmt.Exec(username, string(passwordHash), passwordSalt)

		if err != nil {
			return errors.New(fmt.Sprintf("Error executing statement: %s", err.Error()))
		}

		fmt.Printf("result was: %s", res) //use result to silence compiler for now.
		return nil
	}

	err := db.ExecuteFn(fn)
	if err != nil {
		fmt.Printf("Error executing db action Users#New: %s \n -- \n", err)
		return outStatus, err
	}
	return outStatus, err
}

// Returns -1 if user count cannot be established.
// Returns 0 or 1 if user exists.
// Can return > 1 but schema would prevent it.
func CountUsers(username string) int {
	var userCount int = -1
	fn := func(database *sql.DB) error {

		stmt, err := database.Prepare(
			"SELECT COUNT(username) FROM users WHERE username = $1")
		if err != nil {
			return errors.New(fmt.Sprintf("error preparing statement: %s", err.Error()))
		}

		record := stmt.QueryRow(username)
		if err != nil {
			return errors.New(fmt.Sprintf("error running statement: %s", err.Error()))
		}

		err = record.Scan(&userCount)
		if err != nil {
			return errors.New(fmt.Sprintf("error returning results %s", err.Error()))
		}

		return nil
	}

	err := db.ExecuteFn(fn)
	if err != nil {
		fmt.Printf("Error executing db action Users#Count: %s \n -- \n", err)
	}

	return userCount
}
