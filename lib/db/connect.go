package db

import (
	lib "babou/lib"

	"database/sql"
	_ "github.com/lib/pq"
)

// A lambda which takes a database connection.
type DbAction func(*sql.DB, error)

// Executes a function.
// Note that this is not thread safe.
// ExecuteFn will only hold the connection open so long as you block.
func ExecuteFn(dba DbAction) {
	db, err := sql.Open("postgres", "user=drbawb dbname=babou sslmode=disable")
	defer db.Close()
	dba(db, err)
}

func ChangeSettings(settings *lib.DbSettings) {
	// alter database settings
}
