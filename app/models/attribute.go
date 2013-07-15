package models

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/drbawb/babou/lib/db"
)

/*var sqlUp string = `CREATE TABLE attributes (
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
)`*/

// An attribute record stores meta-data which can be attached
// to a torrent.
type Attribute struct {
	id        int // Not usable from web-app
	TorrentId int //

	Name        string
	ArtistName  []string
	AlbumName   string
	ReleaseYear time.Time

	MusicFormat string
	DiscNumber  int
	Discs       int

	AlbumDescription   string
	ReleaseDescription string
}

func (elem *Attribute) SelectTorrent(torrentId int) error {

	selectAttributes := `SELECT 
	attribute_id, torrent_id, 
	name, artist_name, album_name, release_year,
	music_format, disc_num, total_discs,
	album_description,release_description
	FROM attributes WHERE torrent_id = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectAttributes, torrentId)
		var artistBytes []byte

		err := row.Scan(&elem.id, &elem.TorrentId,
			&elem.Name, &artistBytes, &elem.AlbumName, &elem.ReleaseYear,
			&elem.MusicFormat, &elem.DiscNumber, &elem.Discs,
			&elem.AlbumDescription, &elem.ReleaseDescription,
		)

		fmt.Printf("len artistNames bytes: %d \n", len(artistBytes))

		if err != nil {
			return err
		}

		artistBuf := bytes.NewBuffer(artistBytes)
		decoder := gob.NewDecoder(artistBuf)

		err = decoder.Decode(&elem.ArtistName)
		if err != nil {
			return err
		}

		return nil
	}

	err := db.ExecuteFn(dba)
	if err != nil {
		return err
	}

	return nil //safe to use returned list
}

func (attributes *Attribute) WriteFor(torrentId int) error {
	insert := `INSERT INTO attributes (
		torrent_id, 
		name, artist_name, album_name, release_year,
		music_format, disc_num, total_discs,
		album_description, release_description
	) VALUES (
		$1, 
		$2, $3, $4, $5, 
		$6, $7, $8,
		$9, $10
	)`

	dba := func(dbConn *sql.DB) error {
		insertStmt, err := dbConn.Prepare(insert)
		if err != nil {
			return err
		}

		encodeBuf := bytes.NewBuffer(make([]byte, 0))
		encoder := gob.NewEncoder(encodeBuf)

		encoder.Encode(attributes.ArtistName)

		_, err = insertStmt.Exec(torrentId,
			attributes.Name, encodeBuf.Bytes(), attributes.AlbumName, attributes.ReleaseYear,
			attributes.MusicFormat, attributes.DiscNumber, attributes.Discs,
			attributes.AlbumDescription, attributes.ReleaseDescription,
		)

		if err != nil {
			return err
		}

		return nil

	}

	return db.ExecuteFn(dba)
}
