package models

import (
	db "babou/lib/db"
	sql "database/sql"
	errors "errors"
	fmt "fmt"
)

// `User` model for `users`
type User struct {
	Username string

	passwordHash string
	passwordSalt string
}

func NewUser(username string, passwordHash, passwordSalt []byte) error {
	fn := func(database *sql.DB) error {
		if CountUsers(username) > 0 {
			fmt.Printf("username already exists")
			return errors.New("Username already exists")
		}

		stmt, err := database.Prepare("INSERT INTO users(username,passwordhash,passwordsalt) VALUES($1,$2, $3)")

		if err != nil {
			return errors.New(fmt.Sprintf("Error preparing statement: %s", err.Error()))
		}

		res, err := stmt.Exec(username, string(passwordHash), passwordSalt)

		if err != nil {
			return errors.New(fmt.Sprintf("Error executing statement: %s", err.Error()))
		}

		fmt.Printf("result was: %s", res) //use result to silence compiler
		return nil
	}

	err := db.ExecuteFn(fn)
	if err != nil {
		fmt.Printf("Error executing db action Users#New: %s \n -- \n", err)
	}
	return nil
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
