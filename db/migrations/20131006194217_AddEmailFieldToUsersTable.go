package main

import (
	"database/sql"
	"fmt"
)

var sqlUp string = `
	ALTER TABLE users
	ADD COLUMN email character varying(255);
`

var sqlDown string = `
	ALTER TABLE users
	DROP COLUMN email;
`

// Up is executed when this migration is applied
func Up_20131006194217(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20131006194217(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}
