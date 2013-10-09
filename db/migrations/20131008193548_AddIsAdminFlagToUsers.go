package main

import (
	"database/sql"
	"fmt"
)

var sqlUp string = `
	ALTER TABLE users
	ADD COLUMN is_admin boolean DEFAULT false;
`

var sqlDown string = `
	ALTER TABLE users
	DROP COLUMN is_admin;
	
`

// Up is executed when this migration is applied
func Up_20131008193548(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20131008193548(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}
