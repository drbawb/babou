package main

import (
	"database/sql"
	"fmt"
)

var sqlUp string = `ALTER TABLE "users" ADD COLUMN "secret" bytea,
ADD COLUMN "secret_hash" bytea`

var sqlDown string = `ALTER TABLE "users" DROP COLUMN "secret", DROP COLUMN "secret_hash"`

// Up is executed when this migration is applied
func Up_20130708104836(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)

	if err != nil {
		fmt.Printf("error adding col announce key: %s", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20130708104836(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)

	if err != nil {
		fmt.Printf("error adding col announce key: %s", err.Error())
	}
}
