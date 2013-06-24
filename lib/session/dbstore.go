package session

// An implementation of  the gorilla/sessions#Store interface for `babou`
//
// This uses a PostgreSQL database as a storage backend.
// This allows sessions to be accessed by any instance of `babou`
// which share a common session encryption key in their configuration.

import (
	securecookie "github.com/gorilla/securecookie"
	sessions "github.com/gorilla/sessions"

	errors "errors"
	http "net/http"

	dbLib "babou/lib/db"
	sql "database/sql"

	base32 "encoding/base32"
	fmt "fmt"
	strings "strings"
)

const (
	SESSIONS_TABLE = "http_sessions"
)

type DatabaseStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	path    string
}

func NewDatabaseStore(keyPairs ...[]byte) *DatabaseStore {
	dbStore := &DatabaseStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
	}

	return dbStore
}

// Fetches a session for a given name after it has been added to the registry.
func (db *DatabaseStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	fmt.Printf("getting session from db . . .")
	return sessions.GetRegistry(r).Get(db, name)
}

// New returns a new session for the given name w/o adding it to the registry.
func (db *DatabaseStore) New(r *http.Request, name string) (*sessions.Session, error) {
	fmt.Printf("entered DbStore#New")
	session := sessions.NewSession(db, name)
	session.Options = &(*db.Options)
	session.IsNew = true

	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, db.Codecs...)
		if err == nil {
			err = db.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}

	return session, err
}

func (db *DatabaseStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.ID == "" {
		// Generate a random session ID key suitable for storage in the DB
		session.ID = string(securecookie.GenerateRandomKey(32))
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}

	if err := db.save(session); err != nil {
		return err
	}

	// Keep the session ID key in a cookie so it can be looked up in DB later.

	fmt.Printf("Saving session w/ ID: %s\n", session.ID)

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, db.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

//load fetches a session by ID from the database and decodes its content into session.Values
func (db *DatabaseStore) load(session *sessions.Session) error {
	fn := func(dbConn *sql.DB) error {
		// Write record to sessions table.
		tx, err := dbConn.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO \""+SESSIONS_TABLE+"\" (key, data) VALUES($1,$2)",
			session.ID, encoded)
		if err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		return nil
	}

}

// save writes encoded session.Values to a database record.
// writes to http_sessions table by default.
func (db *DatabaseStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		db.Codecs...)

	if err != nil {
		return err
	}

	fn := func(dbConn *sql.DB) error {
		// Write record to sessions table.
		tx, err := dbConn.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO \""+SESSIONS_TABLE+"\" (key, data) VALUES($1,$2)",
			session.ID, encoded)
		if err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		return nil
	}

	if err = dbLib.ExecuteFn(fn); err != nil {
		fmt.Printf("dbLib executeFn: %s", err.Error())
		return err
	}

	return nil
}
