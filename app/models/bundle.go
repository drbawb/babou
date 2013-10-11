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
	Name     string
	Episodes []*EpisodeBundle
}

type EpisodeBundle struct {
	Number int
	Name   string

	Format     string
	Resolution string
}

func LatestSeries() []*SeriesBundle {
	seriesByID := make(map[int]*SeriesBundle)
	seriesList := make([]*SeriesBundle, 0)

	loadSeriesBundles := `
	SELECT
		episode.parent_id, series.attributes_bundle_id, series.bundle, episode.bundle
	FROM attributes_bundle AS series
	INNER JOIN attributes_bundle AS episode
		ON series.attributes_bundle_id = episode.parent_id
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
			var seriesId, episodeId int
			var seriesBundle, episodeBundle hstore.Hstore

			err := rows.Scan(&seriesId, &episodeId, &seriesBundle, &episodeBundle)
			if err != nil {
				return err
			}

			episode := &EpisodeBundle{}
			episode.FromBundle(episodeBundle.Map)

			if series, ok := seriesByID[seriesId]; ok {
				series.Episodes = append(series.Episodes, episode)
			} else {
				series := &SeriesBundle{Episodes: make([]*EpisodeBundle, 0)}
				series.FromBundle(seriesBundle.Map)
				seriesByID[seriesId] = series

				seriesList = append(seriesList, series)
				series.Episodes = append(series.Episodes, episode)
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
