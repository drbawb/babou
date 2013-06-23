package models

import (
	db "babou/lib/db"
	sql "database/sql"
	fmt "fmt"
)

// `User` model for `users`
type User struct {
	Username string

	passwordHash string
	passwordSalt string
}

func NewUser(username string, passwordHash, passwordSalt []byte) error {
	fmt.Printf("new user . . .")

	fn := func(database *sql.DB, err error) {
		if CountUsers(username) > 0 {
			fmt.Printf("username already exists")
			return
		}

		if err != nil {
			fmt.Printf("database error")
			return
		}

		stmt, err := database.Prepare("INSERT INTO users(username,passwordhash,passwordsalt) VALUES($1,$2, $3)")

		if err != nil {
			fmt.Printf("error preparing statement: %s", err.Error())
		}

		res, err := stmt.Exec(username, string(passwordHash), passwordSalt)

		if err != nil {
			fmt.Printf("query error: %s", err.Error())
			return
		}

		fmt.Printf("result was: %s", res)

	}

	db.ExecuteFn(fn)

	return nil
}

// Returns -1 if user count cannot be established.
// Returns 0 or 1 if user exists.
// Can return > 1 but schema would prevent it.
func CountUsers(username string) int {
	fmt.Printf("checking for users...")
	var userCount int = -1
	fn := func(database *sql.DB, err error) {
		if err != nil {
			fmt.Printf("database error")
			return
		}

		stmt, err := database.Prepare(
			"SELECT COUNT(username) FROM users WHERE username = $1")
		if err != nil {
			fmt.Printf("error preparing statement")
			return
		}

		record := stmt.QueryRow(username)
		if err != nil {
			fmt.Printf("error running statement: %s", err.Error())
			return
		}

		err = record.Scan(&userCount)
		if err != nil {
			fmt.Printf("error returning results %s", err.Error())
			return
		}
	}
	db.ExecuteFn(fn)
	return userCount
}
