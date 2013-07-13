package main

import (
	"database/sql"
	"fmt"
)

// An attribute record stores the necessary metadata for a torrent.
var sqlUp string = `CREATE TABLE attributes (
	  attribute_id serial NOT NULL,
	  torrent_id integer REFERENCES torrents(torrent_id),
	  
	  name character varying(60) NOT NULL,
	  artist_name character varying(255)[],
	  album_name character varying(255),
	  release_year date,
	  
	  music_format character varying(24),
	  disc_num integer,
	  total_discs integer,

	  album_description text,
	  release_description text,

	  CONSTRAINT attributes_pkey PRIMARY KEY (attribute_id)
)`

var sqlDown string = `DROP TABLE attributes`

// Up is executed when this migration is applied
func Up_20130713085811(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20130713085811(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("error commiting txn: %s\n", err.Error())
	}

}
