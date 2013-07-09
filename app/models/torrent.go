package models

import (
	sql "database/sql"
	db "github.com/drbawb/babou/lib/db"

	"github.com/drbawb/babou/lib/torrent"

	"fmt"
)

type Torrent struct {
	ID       int
	Name     string
	InfoHash string

	CreatedBy    string
	CreationDate int

	Encoding    string
	EncodedInfo []byte

	isInit bool
}

/*dba := func(dbConn *sql.DB) error {
err := db.ExecuteFn(dba)*/

// Looks up a torrent based on its info hash,
// this is a 20-byte SHA which is encoded as a string [2-chars per byte.]
func (t *Torrent) SelectHash(hash string) error {
	selectTorrent := `SELECT 
	torrent_id, name, info_hash, created_by, creation_date, encoding, info_bencoded
	FROM "torrents"
		WHERE "info_hash" = $1`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(selectTorrent, hash)
		err := row.Scan(&t.ID, &t.Name, &t.InfoHash, &t.CreatedBy, &t.CreationDate,
			&t.Encoding, &t.EncodedInfo)

		if err == nil {
			t.isInit = true
		}

		return err
	}

	return db.ExecuteFn(dba)
}

func (t *Torrent) Write() error {
	insertTorrent := `INSERT INTO torrents 
	(name, info_hash, created_by, creation_date, encoding, info_bencoded)
	VALUES ($1,$2,$3,$4,$5,$6) RETURNING torrent_id`

	updateTorrent := `UPDATE torrents
		SET name = $2
		SET info_hash = $3
		SET created_by = $4
		SET creation_date = $5
		SET encoding = $6
		SET info_bencoded = $7
	WHERE torrent_id = $1`

	dba := func(dbConn *sql.DB) error {
		noRowsUpdated := true

		//try update then insert
		if t.ID > 0 {
			res, err := dbConn.Exec(updateTorrent, t.ID, t.Name, t.InfoHash,
				t.CreatedBy, t.CreationDate, t.Encoding, t.EncodedInfo)
			if err != nil {
				return err
			}

			rowsAffected, err := res.RowsAffected()
			if err != nil {
				return err
			}

			if rowsAffected > 0 {
				noRowsUpdated = false
			}
		}

		//row not updated; do an insert.
		if noRowsUpdated {
			fmt.Printf("performing insert for torrent \n")
			err := dbConn.QueryRow(insertTorrent, t.Name, t.InfoHash,
				t.CreatedBy, t.CreationDate, t.Encoding, t.EncodedInfo).Scan(&t.ID)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return db.ExecuteFn(dba)

}

func (t *Torrent) Populate(torrentFile *torrent.TorrentFile) error {
	t.CreatedBy = torrentFile.CreatedBy
	t.CreationDate = int(torrentFile.CreationDate)
	t.Encoding = torrentFile.Encoding
	t.InfoHash = fmt.Sprintf("%x", torrentFile.EncodeInfo())

	encodedInfo, err := torrentFile.BencodeInfoDict()
	if err != nil {
		fmt.Printf("error writing encoded info file.")
		return err
	}

	t.EncodedInfo = encodedInfo
	t.isInit = true

	return nil
}
