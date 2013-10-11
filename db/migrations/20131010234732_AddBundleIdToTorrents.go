package main

import (
	"database/sql"
	"fmt"
)

// An attribute record stores the necessary metadata for a torrent.
var sqlUp string = `
ALTER TABLE torrents
ADD COLUMN attributes_bundle_id integer REFERENCES attributes_bundle(attributes_bundle_id)
`

var sqlDown string = `
ALTER TABLE torrents
DROP COLUMN attributes_bundle_id
`

// Up is executed when this migration is applied
func Up_20131010234732(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20131010234732(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}

}
