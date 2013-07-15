package main

import (
	"database/sql"
	"fmt"
)

var sqlUp string = `
	ALTER TABLE attributes
	DROP COLUMN artist_name;

	ALTER TABLE attributes
	ADD COLUMN artist_name bytea;
`

var sqlDown string = `
	ALTER TABLE attributes
	DROP COLUMN artist_name;
	ADD COLUMN artist_name character varying(60)[]
`

// Up is executed when this migration is applied
func Up_20130713094315(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20130713094315(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}
