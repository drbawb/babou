package main

import (
	"database/sql"
)

var createTable string = `
CREATE TABLE sessions
(
	  session_id serial NOT NULL,
	  user_id integer NOT NULL,
	  login_ip inet,
	  CONSTRAINT sessions_pkey PRIMARY KEY (session_id),
	  CONSTRAINT sessions_users_fkey FOREIGN KEY (user_id) REFERENCES "users"
)`

var dropTable string = `
DROP TABLE "sessions"`

// Up is executed when this migration is applied
func Up_20130624095937(txn *sql.Tx) {
	txn.Exec(createTable)
}

// Down is executed when this migration is rolled back
func Down_20130624095937(txn *sql.Tx) {
	txn.Exec(dropTable)
}
