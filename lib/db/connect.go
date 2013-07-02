// Exposes a connection to the underlying PostgreSQL RDBMS.
//
// This library federates access to the database and provides a common point for
// configuration as well as optimzation.
package db

import (
	errors "errors"

	lib "github.com/drbawb/babou/lib"

	"database/sql"
	_ "github.com/lib/pq"
)

var currentConn *babouDb = nil

type babouDb struct {
	database *sql.DB
	msgIO    chan<- DbMessage
}

type DbMessage int

const (
	CLOSE_DB DbMessage = iota
)

// A lambda which takes an open database connection.
// The lambda can return a friendly model-level error which will be passed
// through ExecuteFn
type DbAction func(*sql.DB) error

// Opens a database connection if one has not been established.
// The channel can be used to forward private requests to the database.
func Open() (<-chan DbMessage, error) {
	if currentConn != nil {
		return nil, errors.New("A database connection is already open.")
	}

	dbConn, err := sql.Open("postgres", "user=rstraw host=localhost dbname=babou sslmode=disable")
	if err != nil {
		return nil, err
	}

	dbMsgChan := make(chan DbMessage, 0)

	currentConn = &babouDb{database: dbConn, msgIO: dbMsgChan}

	return dbMsgChan, nil
}

// Executes a function.
// Note that this is not thread safe.
// ExecuteFn will only hold the connection open so long as you block.
func ExecuteFn(dba DbAction) error {
	if currentConn == nil {
		return errors.New("There is no open database connection.")
	}

	// TODO: How come auth-failure doesn't happen until I try to prepare a statement?

	dbaErr := dba(currentConn.database)
	return dbaErr
}

func ChangeSettings(settings *lib.DbSettings) {
	// alter database settings
}
