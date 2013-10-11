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
	TorrentID int `json:"torrentId"`

	Name     string           `json:"name"`
	Episodes []*EpisodeBundle `json:"episodes"`
}

type EpisodeBundle struct {
	TorrentID int `json:"torrentId"`

	Number int    `json:"number"`
	Name   string `json:"name"`

	Format     string `json:"format"`
	Resolution string `json:"resolution"`
}

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
			var seriesIdN64 sql.NullInt64
			var seriesId, episodeId, torrentId int
			var seriesBundle, episodeBundle hstore.Hstore

			err := rows.Scan(&seriesIdN64, &episodeId, &torrentId, &seriesBundle, &episodeBundle)
			if err != nil {
				return err
			}

			seriesId = int(seriesIdN64.Int64) // TODO: loss of precision.

			// Create a map-entry for the series if we haven't seen it yet.
			if _, ok := seriesByID[seriesId]; !ok {
				series := &SeriesBundle{Episodes: make([]*EpisodeBundle, 0)}
				series.TorrentID = torrentId
				series.FromBundle(seriesBundle.Map)
				seriesByID[seriesId] = series

				seriesList = append(seriesList, series)
			}

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
	return len(sb.Episodes)
}
