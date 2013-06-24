package main

import (
	"database/sql"
)

var createTable string = `
CREATE TABLE http_sessions
(
	  http_session_id serial NOT NULL,
	  key bytea,
	  data text,
	  CONSTRAINT http_sessions_pkey PRIMARY KEY (http_session_id)
)`

var dropTable string = `
DROP TABLE "http_sessions"`

// Up is executed when this migration is applied
func Up_20130624123610(txn *sql.Tx) {
	txn.Exec(createTable)
}

// Down is executed when this migration is rolled back
func Down_20130624123610(txn *sql.Tx) {
	txn.Exec(dropTable)
}
