package models

import (
	"database/sql"
	"github.com/drbawb/babou/lib/db"
	"github.com/pbnjay/pq/hstore"

	"log"
	"strconv"
	"time"
)

type BundleType string

const (
	BUNDLE_SERIES  BundleType = "series"
	BUNDLE_EPISODE            = "episode"
)

type Bundle interface {
	ToBundle() map[string]string
	FromBundle(map[string]string) error
}

// This is my bundle
// There are many like it,
// but this one is mine.
type AttributesBundle struct {
	ID       int
	ParentID int

	Category BundleType
	Bundle   hstore.Hstore
	Modified time.Time
}

type SeriesBundle struct {
	ID        int `json:"bundleId"`
	TorrentID int `json:"torrentId"`

	Name     string           `json:"name"`
	Episodes []*EpisodeBundle `json:"episodes"`
}

type EpisodeBundle struct {
	ID        int `json:"episodeId"`
	TorrentID int `json:"torrentId"`

	Number int    `json:"number"`
	Name   string `json:"name"`

	Format     string `json:"format"`
	Resolution string `json:"resolution"`
}

// Attempts to populate series bundle by name.
// Otherwise this series bundle will be inserted.
func (sb *SeriesBundle) SelectByName(seriesName string) error {
	sbByName := `
	SELECT 
		bundle
	FROM 
		attributes_bundle
	WHERE category = ?
	AND bundle->'name' = ?
	`

	dba := func(dbConn *sql.DB) error {
		row := dbConn.QueryRow(sbByName, BUNDLE_SERIES, seriesName)

		var bundleStore hstore.Hstore
		if err := row.Scan(&bundleStore); err != nil {
			return err
		}

		if err := sb.FromBundle(bundleStore.Map); err != nil {
			return err
		}

		return nil
	}

	if err := db.ExecuteFn(dba); err != nil {
		// Reinit; they can do whatever I don't care.
		sb = &SeriesBundle{}
		return err
	} else {
		return nil
	}
}

func (sb *SeriesBundle) Persist() error {
	sbInsert := `
	INSERT INTO 
		attributes_bundle(category, bundle, modified) 
	VALUES 
		($1, $2, $3) 
	RETURNING attributes_bundle_id`

	dba := func(dbConn *sql.DB) error {
		hBundle := &hstore.Hstore{Map: sb.ToBundle()}

		row := dbConn.QueryRow(sbInsert, string(BUNDLE_SERIES), hBundle, time.Now())
		if err := row.Scan(&sb.ID); err != nil {
			return err
		}

		return nil
	}

	return db.ExecuteFn(dba)
}

// TODO: Try update first.
func (sb *SeriesBundle) PersistWith(tx *sql.Tx) error {
	sbInsert := `
	INSERT INTO "attributes_bundle"
		(category, bundle, modified)
	VALUES
		('series', ?, ?)
	RETURNING
		attributes_bundle_id
	`

	row := tx.QueryRow(sbInsert, sb.ToBundle(), time.Now())
	if err := row.Scan(&sb.ID); err != nil {
		return err
	}

	return nil
}

// Selects the latest series' of television.
//
// This will select 100 series at a time which have a torrent or episode(s)
// associated with them.
//
func LatestSeries() []*SeriesBundle {
	seriesByID := make(map[int]*SeriesBundle)
	seriesList := make([]*SeriesBundle, 0)

	// Selects all episodes from series
	// If no episodes are avail., the series itself is selected.
	// (This would happen if, for e.g, the series is related to a multi-file torrent.)
	loadSeriesBundles := `
	SELECT
		episode.parent_id, series.attributes_bundle_id, tor.torrent_id, series.bundle, episode.bundle
	FROM attributes_bundle AS series
	LEFT JOIN attributes_bundle AS episode
		ON series.attributes_bundle_id = episode.parent_id
	INNER JOIN torrents AS tor
		ON tor.attributes_bundle_id = episode.attributes_bundle_id
		OR tor.attributes_bundle_id = series.attributes_bundle_id
	WHERE series.category = 'series'
	ORDER BY series.modified DESC
	LIMIT 100
	`

	dba := func(dbConn *sql.DB) error {
		rows, err := dbConn.Query(loadSeriesBundles)
		if err != nil {
			return err
		}

		for rows.Next() {
			var epIdN, serIdN sql.NullInt64
			var seriesId, episodeId, torrentId int
			var seriesBundle, episodeBundle hstore.Hstore

			err := rows.Scan(
				&epIdN,
				&serIdN,
				&torrentId,
				&seriesBundle,
				&episodeBundle)

			if err != nil {
				return err
			}

			seriesId = int(serIdN.Int64) // TODO: loss of precision.
			episodeId = int(epIdN.Int64)

			// Create a map-entry for the series if we haven't seen it yet.
			// Add series if not exist

			var seriesIdx int
			if !epIdN.Valid {
				// This is a series, set the series index
				// to the series' attribute_bundle identifier
				seriesIdx = seriesId
			} else {
				// This is an episode, set the series index
				// to the episodes' parent ID.
				// (The series' attribute_bundle identifier.)
				seriesIdx = episodeId
			}

			// If we haven't seen this series before
			// add it to our in-mem structure
			if _, ok := seriesByID[seriesIdx]; !ok {
				series := &SeriesBundle{Episodes: make([]*EpisodeBundle, 0)}
				series.TorrentID = torrentId
				series.FromBundle(seriesBundle.Map)
				seriesByID[seriesId] = series

				seriesList = append(seriesList, series)
			}

			// If this has an episodes' bundle associated with it
			// load that bundle into its related series.
			if episodeBundle.Map != nil {
				// Attach episode bundle if applicable.
				episode := &EpisodeBundle{}

				episode.TorrentID = torrentId
				episode.FromBundle(episodeBundle.Map)

				seriesByID[seriesId].Episodes = append(
					seriesByID[seriesId].Episodes,
					episode,
				)
			}

		}

		return nil
	}

	err := db.ExecuteFn(dba)
	if err != nil {
		log.Printf("Fetching latest series error: %s \n", err.Error())
	}

	return seriesList
}

func (eb *EpisodeBundle) PersistWithSeries(series *SeriesBundle) error {

	insertEb := `
	INSERT INTO attributes_bundle
		(parent_id, category, bundle, modified)
	VALUES 
		(?, ?, ?, ?)
	RETURNING
		attributes_bundle_id
	`
	dba := func(dbConn *sql.DB) error {
		txn, err := dbConn.Begin()
		if err != nil {
			return err
		}

		// Persist series and retrieve ID
		if err = series.PersistWith(txn); err != nil {
			_ = txn.Rollback()
			return err
		}

		row := txn.QueryRow(
			insertEb,
			series.ID,
			BUNDLE_EPISODE,
			eb.ToBundle(),
			time.Now())
		err = row.Scan(&eb.ID)
		if err != nil {
			_ = txn.Rollback()
			return err
		}

		if err = txn.Commit(); err != nil {
			return err
		}

		return nil
	}

	return db.ExecuteFn(dba)

}

// TODO: Pagination
// Select the latest episodes, regardless of series.
func LatestEpisodes() []*EpisodeBundle {
	episodes := make([]*EpisodeBundle, 0)

	loadBundles := `
	SELECT 
		bundle
	FROM attributes_bundle
	WHERE category = 'episode'
	ORDER BY modified DESC
	LIMIT 100
	`

	dba := func(dbConn *sql.DB) error {
		rows, err := dbConn.Query(loadBundles)
		if err != nil {
			return err
		}

		for rows.Next() {
			// Scan hstore bundle
			var mappedBundle hstore.Hstore
			eb := &EpisodeBundle{}

			if err := rows.Scan(&mappedBundle); err != nil {
				return err
			}

			// Parse hstore into episode bundle
			if err := eb.FromBundle(mappedBundle.Map); err != nil {
				return err
			}

			// Append episode bundle to final list
			episodes = append(episodes, eb)
		}

		return nil
	}

	// Trouble fetching bundles?
	err := db.ExecuteFn(dba)
	if err != nil {
		log.Printf("Error fetching latest episodes: %s \n", err.Error())
	}

	return episodes
}

func (eb *EpisodeBundle) ToBundle() map[string]sql.NullString {
	bundleMap := make(map[string]sql.NullString)

	bundleMap["name"] = sql.NullString{String: eb.Name}
	bundleMap["number"] = sql.NullString{String: strconv.Itoa(eb.Number)}
	bundleMap["format"] = sql.NullString{String: eb.Format}
	bundleMap["resolution"] = sql.NullString{String: eb.Resolution}

	return bundleMap
}

// Implements Bundle interface and loads an EpisodeBundle
// object from a deserialized postgresql hstore field.
func (eb *EpisodeBundle) FromBundle(bundleStore map[string]sql.NullString) error {
	var err error

	eb.Name = bundleStore["name"].String
	eb.Number, err = strconv.Atoi(bundleStore["number"].String)
	if err != nil {
		return err
	}

	eb.Format = bundleStore["format"].String
	eb.Resolution = bundleStore["resolution"].String

	return nil
}

func (sb *SeriesBundle) ToBundle() map[string]sql.NullString {
	bundleMap := make(map[string]sql.NullString)

	bundleMap["name"] = sql.NullString{sb.Name, true}

	return bundleMap
}

// Implements Bundle interface and loads an SeriesBundle
// object from a deserialized postgresql hstore field.
func (sb *SeriesBundle) FromBundle(bundleStore map[string]sql.NullString) error {
	sb.Name = bundleStore["name"].String

	return nil
}

func (sb *SeriesBundle) Head() *EpisodeBundle {
	if len(sb.Episodes) > 0 {
		return sb.Episodes[0]
	}

	return nil
}

func (sb *SeriesBundle) Tail() []*EpisodeBundle {
	if len(sb.Episodes) > 1 {
		return sb.Episodes[1:]
	}

	return nil
}

func (sb *SeriesBundle) NumberOfEpisodes() int {
	if len(sb.Episodes) <= 0 {
		return 1
	} else {
		return len(sb.Episodes)
	}
}
