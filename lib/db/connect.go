// Exposes a connection to the underlying PostgreSQL RDBMS.
//
// This library federates access to the database and provides a common point for
// configuration as well as optimzation.
package db

import (
	errors "errors"
	fmt "fmt"

	lib "github.com/drbawb/babou/lib"

	"database/sql"
	_ "github.com/lib/pq"
)

var currentConn BabouDb = nil

type BabouDb struct {
	database *sql.DB
}

// A lambda which takes an open database connection.
// The lambda can return a friendly model-level error which will be passed
// through ExecuteFn
type DbAction func(*sql.DB) error

// Executes a function.
// Note that this is not thread safe.
// ExecuteFn will only hold the connection open so long as you block.
func ExecuteFn(dba DbAction) error {
	if currentConn == nil {
		db, err := sql.Open("postgres", "user=rstraw host=localhost dbname=babou sslmode=disable")
		if err != nil {
			return errors.New(fmt.Sprintf("There was an error opening the database connection: %s", err))
		}

		currentConn = &BabouDb{database: db}
	}

	// TODO: How come auth-failure doesn't happen until I try to prepare a statement?

	dbaErr := dba(currentConn.database)
	return dbaErr
}

// Will
func FlushPool() error {
	if currentConn != nil {
		err := currentConn.database.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func ChangeSettings(settings *lib.DbSettings) {
	// alter database settings
}
