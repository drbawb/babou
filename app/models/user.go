package models

import (
	sql "database/sql"
	errors "errors"
	fmt "fmt"

	bcrypt "code.google.com/p/go.crypto/bcrypt"
	rand "crypto/rand"

	db "github.com/drbawb/babou/lib/db"
)

// `User` model for `users`
type User struct {
	UserId   int
	Username string

	passwordHash string
	passwordSalt string

	isInit bool
}

// Error-codes returned from some methods that could be presented to the UI.
type UserModelError int

const (
	USERNAME_TAKEN        UserModelError = 1 << iota
	USERNAME_INVALID_CHAR                = 1 << iota
)

// Select user by ID number and populate the current `user` struct with the record data.
// Returns an error if there was a problem. fetching the user information from the database.
func (u *User) SelectId(id int) error {
	selectUserById := `SELECT user_id, username, passwordhash, passwordsalt
	FROM "users" WHERE user_id = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectUserById, id)
		err := row.Scan(&u.UserId, &u.Username, &u.passwordHash, &u.passwordSalt)
		if err != nil {
			return err
		}

		u.isInit = true
		return nil
	}

	err := db.ExecuteFn(dba)
	if err != nil {
		return err
	}

	return nil //safe to use pointer.
}

// Select user by username and populate the current `user` struct with the record data.
// Returns an error if there was a problem. fetching the user information from the database.
func (u *User) SelectUsername(username string) error {
	selectUserById := `SELECT user_id, username, passwordhash, passwordsalt
	FROM "users" WHERE username = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectUserById, username)
		err := row.Scan(&u.UserId, &u.Username, &u.passwordHash, &u.passwordSalt)
		if err != nil {
			return err
		}

		u.isInit = true
		return nil
	}

	err := db.ExecuteFn(dba)
	if err != nil {
		return err
	}

	return nil //safe to use pointer.
}

// Must be performed on an initialized user-struct.
// Returns an error if the user's password cannot be validated.
func (u *User) CheckHash(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(u.passwordSalt+password))
	if err != nil {
		return errors.New("The password you entered is incorrect. Please try again. You have [n] tries remaining.")
	}

	return err
}

// Creates a new user record in the database.
// The first return parameter is an error-code that represents a non-fatal
// problem that could be presented to the user.
//
// The second return parameter is a fatal error passed up from the database layer.
func NewUser(username, password string) (UserModelError, error) {
	var outStatus UserModelError

	fn := func(database *sql.DB) error {
		if CountUsers(username) > 0 {
			outStatus = USERNAME_TAKEN
			return nil
		}

		passwordHash, passwordSalt, err := genHash(password)
		if err != nil {
			return err
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

// Takes a password and returns a hash and salt.
func genHash(password string) (string, []byte, error) {
	//64-char salt
	saltLength := 64
	passwordSalt := make([]byte, saltLength)
	saltedPassword := make([]byte, 0)

	n, err := rand.Read(passwordSalt)
	if n != len(passwordSalt) || err != nil {
		return "", nil, errors.New("Error generating salt for password.")
	}

	saltedPassword = append(saltedPassword, []byte(password)...)
	hashedPassword, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.MinCost)

	if err != nil {
		return "", nil, errors.New("Error encrypting password for storage.")
	}

	return string(hashedPassword), passwordSalt, nil
}
