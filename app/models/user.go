package models

import (
	sql "database/sql"
	errors "errors"
	fmt "fmt"

	bcrypt "code.google.com/p/go.crypto/bcrypt"

	"encoding/hex"

	hmac "crypto/hmac"
	rand "crypto/rand"
	sha256 "crypto/sha256"

	db "github.com/drbawb/babou/lib/db"
)

// `User` model for `users`
type User struct {
	UserId   int
	Username string
	IsAdmin  bool

	Email    string
	emailSql sql.NullString

	passwordHash string
	passwordSalt string

	Secret     []byte
	SecretHash []byte

	isInit bool
}

// Error-codes returned from some methods that could be presented to the UI.
type UserModelError int

const (
	USERNAME_TAKEN        UserModelError = 1 << iota
	USERNAME_INVALID_CHAR                = 1 << iota
	FAIL_GEN_SECRET                      = 1 << iota
)

func AllUsers() ([]*User, error) {
	usersList := make([]*User, 0)
	selectUsers := `SELECT user_id, username, email, passwordhash, passwordsalt, secret, secret_hash
	FROM "users"`

	dba := func(dbConn *sql.DB) error {
		rows, err := dbConn.Query(selectUsers)
		if err != nil {
			return err
		}

		for rows.Next() {
			u := &User{}
			err := rows.Scan(
				&u.UserId,
				&u.Username,
				&u.emailSql,
				&u.passwordHash,
				&u.passwordSalt,
				&u.Secret,
				&u.SecretHash)

			if err != nil {
				return err
			}

			u.Email = u.emailSql.String

			usersList = append(usersList, u)
		}

		return nil
	}

	err := db.ExecuteFn(dba)
	if err != nil {
		return usersList, err
	}

	return usersList, err //safe to use pointer.
}

// Select user by ID number and populate the current `user` struct with the record data.
// Returns an error if there was a problem. fetching the user information from the database.
func (u *User) SelectId(id int) error {
	selectUserById := `SELECT user_id, username, is_admin, passwordhash, passwordsalt, secret, secret_hash
	FROM "users" WHERE user_id = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectUserById, id)
		err := row.Scan(&u.UserId, &u.Username, &u.IsAdmin, &u.passwordHash, &u.passwordSalt, &u.Secret, &u.SecretHash)
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

func (u *User) Delete() error {
	deleteUserById := `DELETE FROM "users" WHERE user_id = $1`
	dba := func(dbConn *sql.DB) error {
		_, err := dbConn.Exec(deleteUserById, u.UserId)
		if err != nil {
			return err
		}

		return nil
	}

	return db.ExecuteFn(dba)
}

// Select user by username and populate the current `user` struct with the record data.
// Returns an error if there was a problem. fetching the user information from the database.
func (u *User) SelectUsername(username string) error {
	selectUserById := `SELECT user_id, username, passwordhash, passwordsalt, secret, secret_hash
	FROM "users" WHERE username = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectUserById, username)
		err := row.Scan(&u.UserId, &u.Username, &u.passwordHash, &u.passwordSalt, &u.Secret, &u.SecretHash)
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

// Selects a user by their secret key. This is used by the tracker
// to authorize a user.
//
// The secret is expected to be a UTF8 string representing a byte array
// using 2-characters per byte. (As per the standard encoding/hex package.)
func (u *User) SelectSecret(secret string) error {
	selectUserBySecret := `SELECT user_id,username,passwordhash,passwordsalt,secret,secret_hash
	FROM "users" WHERE secret = $1`

	secretHex, err := hex.DecodeString(secret)
	if err != nil {
		return errors.New("The user's secret was not in the expected format.")
	}

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectUserBySecret, secretHex)
		err := row.Scan(&u.UserId, &u.Username, &u.passwordHash, &u.passwordSalt, &u.Secret, &u.SecretHash)
		if err != nil {
			return err
		}

		u.isInit = true
		return nil
	}

	err = db.ExecuteFn(dba)
	if err != nil {
		return err
	}

	return nil //safe to use pointer.

}

// Must be performed on an initialized user-struct.
// Returns an error if the user's password cannot be validated.
func (u *User) CheckHash(password string) error {
	saltedPassword := append(make([]byte, 0), []byte(u.passwordSalt+password)...)

	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), saltedPassword)
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

		announceSecret, announceHash, err := genSecret()
		if err != nil {
			return err
		}

		stmt, err := database.Prepare("INSERT INTO users(username,passwordhash,passwordsalt,secret,secret_hash) VALUES($1,$2, $3, $4, $5)")
		fmt.Printf("user registered; [DEBUG] \n secret: %s \n hash: %s \n\n", fmt.Sprintf("%x", announceSecret), fmt.Sprintf("%x", announceHash))

		if err != nil {
			return errors.New(fmt.Sprintf("Error preparing statement: %s", err.Error()))
		}

		res, err := stmt.Exec(username, string(passwordHash), passwordSalt, announceSecret, announceHash)

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

//TODO: needs to use sites base url.
func (u *User) AnnounceURL() string {
	outStr := ""

	outStr = fmt.Sprintf("http://tracker.fatalsyntax.com:4200/%s/%s/announce", hex.EncodeToString(u.Secret), hex.EncodeToString(u.SecretHash))

	return outStr
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

	saltedPassword = append(saltedPassword, passwordSalt...)
	saltedPassword = append(saltedPassword, []byte(password)...)
	hashedPassword, err := bcrypt.GenerateFromPassword(saltedPassword, 0)

	if err != nil {
		return "", nil, err
	}

	return string(hashedPassword), passwordSalt, nil
}

// Generates a random announce-secret.
// The secret is stored in the user's profile
//
// This secret is appended to the database along with a hash
//
// HMAC to verify the authenticity of the secret; and use the secret
// to lookup the user.
// Returns: secret, secret's hash, and any error encountered.
func genSecret() ([]byte, []byte, error) {
	//TODO: store shared tracker secret in configuration.
	sharedKey := []byte("f75778f7425be4db0369d09af37a6c2b9ab3dea0e53e7bd57412e4b060e607f7")

	randomSecret := make([]byte, 64)
	n, err := rand.Read(randomSecret)
	if err != nil {
		return nil, nil, err
	} else if n != len(randomSecret) {
		return nil, nil, errors.New("Error generating random secret for user.")
	}

	// The secret and its hash will be sent with each tracker request
	mac := hmac.New(sha256.New, sharedKey)
	mac.Write(randomSecret)

	return randomSecret, mac.Sum(nil), nil
}
