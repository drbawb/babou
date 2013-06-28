// An implementation of  the gorilla/sessions#Store interface for `babou`
//
// This uses a PostgreSQL database as a storage backend.
// This allows sessions to be accessed by any instance of `babou`
// which share a common session encryption key in their configuration.
package session

import (
	securecookie "github.com/gorilla/securecookie"
	sessions "github.com/gorilla/sessions"

	http "net/http"

	sql "database/sql"
	lib "github.com/drbawb/babou/lib"
	dbLib "github.com/drbawb/babou/lib/db"

	base32 "encoding/base32"
	strings "strings"

	errors "errors"
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
	return sessions.GetRegistry(r).Get(db, name)
}

// New returns a new session for the given name w/o adding it to the registry.
func (db *DatabaseStore) New(r *http.Request, name string) (*sessions.Session, error) {
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
		lib.Printf("Loading session id [%s] from database.\n", session.ID)
		row := dbConn.QueryRow("SELECT http_session_id, key, data FROM \""+SESSIONS_TABLE+"\" WHERE key = $1", session.ID)

		var id int
		var key, data string
		if err := row.Scan(&id, &key, &data); err != nil {
			return err
		}

		if err := securecookie.DecodeMulti(session.Name(), string(data),
			&session.Values, db.Codecs...); err != nil {
			return err
		}

		return nil
	}

	return dbLib.ExecuteFn(fn)
}

// save writes encoded session.Values to a database record.
// writes to http_sessions table by default.
func (db *DatabaseStore) save(session *sessions.Session) error {
	lib.Printf("Saving session id [%s] to database.\n", session.ID)
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		db.Codecs...)

	if err != nil {
		return err
	}

	fn := func(dbConn *sql.DB) error {
		// Write record to sessions table.
		var sessionCount int = -1

		// Session exists?
		row := dbConn.QueryRow("SELECT COUNT(key) AS count FROM \""+SESSIONS_TABLE+"\" WHERE key = $1", session.ID)

		err := row.Scan(&sessionCount)
		if err != nil {
			return err
		}

		tx, err := dbConn.Begin()
		if err != nil {
			return err
		}

		if sessionCount > 0 {
			// update
			_, err = tx.Exec("UPDATE \""+SESSIONS_TABLE+"\" SET data = $1 WHERE key = $2",
				encoded, session.ID)
			if err != nil {
				return err
			}
		} else if sessionCount == 0 {
			// insert
			_, err = tx.Exec("INSERT INTO \""+SESSIONS_TABLE+"\" (key, data) VALUES($1,$2)",
				session.ID, encoded)
			if err != nil {
				return err
			}
		} else {
			// error
			err = errors.New("There was an error while trying to lookup a previous session.")
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		return nil
	}

	return dbLib.ExecuteFn(fn)
}
