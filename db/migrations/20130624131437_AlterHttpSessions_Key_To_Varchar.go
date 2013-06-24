package main

import (
	"database/sql"
	"fmt"
)

var alterTo string = `
ALTER TABLE http_sessions
ALTER COLUMN key TYPE character varying(60)`

var alterFrom string = `
ALTER TABLE http_sessions
ALTER COLUMN key TYPE bytea`

// Up is executed when this migration is applied
func Up_20130624131437(txn *sql.Tx) {
	txn.Exec(alterTo)
}

// Down is executed when this migration is rolled back
func Down_20130624131437(txn *sql.Tx) {
	fmt.Printf("\nMigration: http_sesions key FROM varchar TO bytea")
	fmt.Printf("\nCannot downcast from `bytea`, please drop and recreate sessions table.")
}
