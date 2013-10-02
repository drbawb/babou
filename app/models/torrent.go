package models

import (
	sql "database/sql"
	db "github.com/drbawb/babou/lib/db"

	"github.com/chuckpreslar/codex"

	"github.com/drbawb/babou/lib/torrent"

	"encoding/hex"
	"errors"
	"fmt"
)

// Define record structure.
// Tags aren't used anywhere . . .
//
// Mostly use them as references, might use them
// for some sort of ORM in the future.
type Torrent struct {
	ID       int    `field:"torrent_id"`
	Name     string `field:"name"`
	InfoHash string `field:"info_hash"`

	CreatedBy    string `field:"created_by"`
	CreationDate int    `field:"creation_date"`

	Encoding    string `field:"encoding"`
	EncodedInfo []byte `field:"info_bencoded"`

	lazyAttributes *Attribute `	table:"attributes" 
								has-one:"torrents" 
								through:"torrent_id"`

	isInit bool
}

/*dba := func(dbConn *sql.DB) error {
err := db.ExecuteFn(dba)*/

func (t *Torrent) SelectId(id int) error {
	// Create a projection
	torrents := codex.Table("torrents")
	torrentsProjection := torrents.Project(
		"torrent_id",
		"name",
		"info_hash",
		"created_by",
		"creation_date",
		"encoding",
		"info_bencoded",
	)

	// Filter results.
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
	torrents := codex.Table("torrents")
	torrentsProjection := torrents.Project(
		"torrent_id",
		"name",
		"info_hash",
		"created_by",
		"creation_date",
		"encoding",
		"info_bencoded",
	)

	torrentsFilter, err := torrentsProjection.Where(
		torrents("info_hash").Eq(hash)).ToSql()
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

// Selects an excerpt of torrents. Only fetches an ID, InfoHash, and CreatedBy
func (t *Torrent) SelectSummaryPage() ([]*Torrent, error) {
	summaryList := make([]*Torrent, 0, 100)

	torrents := codex.Table("torrents")
	torrentsProjection := torrents.Project(
		"torrent_id",
		"name",
		"info_hash",
		"created_by",
	)

	torrentsFilter, err := torrentsProjection.Limit(100).ToSql()
	if err != nil {
		return nil, err
	}

	dba := func(dbConn *sql.DB) error {
		rows, err := dbConn.Query(torrentsFilter)
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
	torrents := codex.Table("torrents")
	torrentsInsert, err := torrents.Insert(
		t.Name,
		t.InfoHash,
		t.CreatedBy,
		t.CreationDate,
		t.Encoding,
		tryHex(t.EncodedInfo),
	).Into(
		"name",
		"info_hash",
		"created_by",
		"creation_date",
		"encoding",
		"info_bencoded",
	).Returning("torrent_id").ToSql()

	if err != nil {
		return err
	}

	updateTorrent, err := torrents.Set(
		"name",
		"info_hash",
		"created_by",
		"creation_date",
		"encoding",
		"info_bencoded",
	).To(
		t.Name,
		t.InfoHash,
		t.CreatedBy,
		t.CreationDate,
		t.Encoding,
		tryHex(t.EncodedInfo),
	).Where(torrents("torrent_id").Eq(t.ID)).ToSql()

	if err != nil {
		return err
	}

	dba := func(dbConn *sql.DB) error {
		noRowsUpdated := true

		//try update then insert
		if t.ID > 0 {
			res, err := dbConn.Exec(updateTorrent)
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
			err := dbConn.QueryRow(torrentsInsert).Scan(&t.ID)
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

// Transforms a []byte into a Postgres hex-escaped string.
// SELECT E'\\xDEADBEEF';
//
// TODO: Endianness is specified by PG spec; ensure encoding/hex complies.
// (The "hex" format encodes binary data as 2 hexadecimal digits per byte, most significant nibble first.)
func encodeInfoForPG(bytea []byte) string {
	hexEscapePrefix := "\\x"
	hexOutBuf := hex.EncodeToString(bytea)

	return hexEscapePrefix + hexOutBuf
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
