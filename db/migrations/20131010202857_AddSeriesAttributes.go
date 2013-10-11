package main

import (
	"database/sql"
	"fmt"
)

// An attribute record stores the necessary metadata for a torrent.
var sqlUp string = `
CREATE TABLE attributes_bundle (
	  attributes_bundle_id serial PRIMARY KEY,
	  parent_id integer,

	  category character varying,
	  bundle hstore,

	  modified timestamp,
	  FOREIGN KEY (parent_id) REFERENCES attributes_bundle(attributes_bundle_id)
)`

var sqlDown string = `DROP TABLE attributes_bundle`

// Up is executed when this migration is applied
func Up_20131010202857(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20131010202857(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}

}
