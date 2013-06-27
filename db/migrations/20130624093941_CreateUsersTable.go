package main

import (
	"database/sql"
)

var createTable string = `
CREATE TABLE users
(
	  user_id serial NOT NULL,
	  username character varying(255),
	  passwordhash character varying(255),
	  passwordsalt bytea,
	  CONSTRAINT users_pkey PRIMARY KEY (user_id )
)`

var dropTable string = `
DROP TABLE "users"`

// Up is executed when this migration is applied
func Up_20130624093941(txn *sql.Tx) {
	txn.Exec(createTable)
}

// Down is executed when this migration is rolled back
func Down_20130624093941(txn *sql.Tx) {
	txn.Exec(dropTable)
}
