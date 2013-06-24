package main

import (
	"database/sql"
)

var alterTo string = `
CREATE UNIQUE INDEX http_sessions_key_idx_unique
  ON http_sessions
  USING btree
  (key COLLATE pg_catalog."default" )`

var alterFrom string = `
DROP INDEX http_sessions_key_idx_unique`

// Up is executed when this migration is applied
func Up_20130624145943(txn *sql.Tx) {
	txn.Exec(alterTo)
}

// Down is executed when this migration is rolled back
func Down_20130624145943(txn *sql.Tx) {
	txn.Exec(alterFrom)
}
