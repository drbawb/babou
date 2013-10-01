package models

import (
	sql "database/sql"
	"github.com/chuckpreslar/codex"
	db "github.com/drbawb/babou/lib/db"

	"github.com/drbawb/babou/lib/torrent"

	"errors"
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

	lazyAttributes *Attribute

	isInit bool
}

/*dba := func(dbConn *sql.DB) error {
err := db.ExecuteFn(dba)*/

func (t *Torrent) SelectId(id int) error {
	torrents := codex.Table("torrents")
	torrentsProjection := torrents.Project("torrent_id", "name", "info_hash", "created_by",
		"creation_date", "encoding", "info_bencoded")
	torrentsFilter, err := torrentsProjection.Where(
		torrents("torrent_id").Eq(id)).ToSql()

	if err != nil {
		return err
	}

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(torrentsFilter)
		err := row.Scan(&t.ID, &t.Name, &t.InfoHash, &t.CreatedBy, &t.CreationDate,
			&t.Encoding, &t.EncodedInfo)

		if err == nil {
			t.isInit = true
		}

		return err
	}

	return db.ExecuteFn(dba)
}

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

// Selects an excerpt of torrents. Only fetches an ID, InfoHash, and CreatedBy
func (t *Torrent) SelectSummaryPage() ([]*Torrent, error) {
	summaryList := make([]*Torrent, 0, 100)

	selectSummary := `SELECT
	torrent_id, name, info_hash, created_by
	FROM "torrents"
	LIMIT 100`

	dba := func(dbConn *sql.DB) error {
		rows, err := dbConn.Query(selectSummary)
		if err != nil {
			return err
		}

		for rows.Next() {
			t := &Torrent{isInit: true}
			_ = rows.Scan(&t.ID, &t.Name, &t.InfoHash, &t.CreatedBy)
			summaryList = append(summaryList, t)
		}

		return nil
	}

	return summaryList, db.ExecuteFn(dba)
}

func (t *Torrent) Attributes() (*Attribute, error) {
	if t.ID <= 0 && t.lazyAttributes == nil {
		return nil, errors.New("This torrent's attributes are not currently available.")
	} else if t.lazyAttributes != nil {
		return t.lazyAttributes, nil
	}

	attributes := &Attribute{}
	err := attributes.SelectTorrent(t.ID)

	return attributes, err

}

func (t *Torrent) SetAttributes(attributes *Attribute) {
	t.lazyAttributes = attributes
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

			if t.lazyAttributes == nil {
				//TODO: attributes not supplied? guess from filenames?
				t.lazyAttributes = &Attribute{}
			}

			err = t.lazyAttributes.WriteFor(t.ID)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return db.ExecuteFn(dba)

}

func (t *Torrent) Populate(torrentFile *torrent.TorrentFile) error {
	if torrentName, ok := torrentFile.Info["name"].(string); ok {
		t.Name = torrentName
	} else {
		return errors.New("Torrent file did not contain a `name` property in the info-hash. Please try recreating your torrent.")
	}
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

func (t *Torrent) WriteFile(secret, hash []byte) ([]byte, error) {
	file, err := t.LoadTorrent()
	if err != nil {
		return nil, err
	}

	outBytes, err := file.WriteFile(secret, hash)
	return outBytes, err
}

func (t *Torrent) LoadTorrent() (*torrent.TorrentFile, error) {
	file := &torrent.TorrentFile{}
	file.Comment = "downloaded from babou development instance"
	file.CreatedBy = t.CreatedBy
	file.CreationDate = int64(t.CreationDate)
	file.Encoding = t.Encoding

	infoMap, err := torrent.DecodeInfoDict(t.EncodedInfo)
	if err != nil {
		return nil, err
	}

	file.Info = infoMap

	return file, nil
}
