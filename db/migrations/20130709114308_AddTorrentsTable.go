package main

import (
	"database/sql"
	"fmt"
)

/**
 Needs to AT LEAST store this:
type TorrentFile struct {
	Announce     string                 `bencode:"announce"`
	Comment      string                 `bencode:"comment"`
	CreatedBy    string                 `bencode:"created by"`
	CreationDate int64                  `bencode:"creation date"`
	Encoding     string                 `bencode:"encoding"`
	Info         map[string]interface{} `bencode:"info"`
}
*/

var sqlUp string = `
CREATE TABLE torrents
(
	  torrent_id serial NOT NULL,
	  name character varying(255) NOT NULL,
	  info_hash character varying(40) NOT NULL,
	  created_by character varying(255),
	  creation_date numeric,
	  encoding character varying(255),
	  info_bencoded bytea,

	  CONSTRAINT torrents_pkey PRIMARY KEY (torrent_id )
)`

var sqlDown string = `
DROP TABLE torrents`

// Up is executed when this migration is applied
func Up_20130709114308(txn *sql.Tx) {
	_, err := txn.Exec(sqlUp)
	if err != nil {
		fmt.Printf("err: %s", err.Error())
	}
}

// Down is executed when this migration is rolled back
func Down_20130709114308(txn *sql.Tx) {
	_, err := txn.Exec(sqlDown)
	if err != nil {
		fmt.Printf("err: %s", err.Error())
	}
}
