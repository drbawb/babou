package models

import (
	"github.com/coopernurse/gorp"

	"database/sql"
	"log"
	"os"
)

var dbMap *gorp.DbMap

func init() {
	dbConn, err := sql.Open("postgres", "dbname=babou host=192.168.1.11 user=drbawb password=babouDev sslmode=disable")
	if err != nil {
		panic(err)
	}

	dbMap = &gorp.DbMap{
		Db:      dbConn,
		Dialect: gorp.PostgresDialect{},
	}

	_ = dbMap.AddTableWithName(User{}, "users").SetKeys(true, "UserId")

	dbMap.TraceOn(
		"[SQL]",
		log.New(
			os.Stdout,
			"babou QUERY:",
			log.Lmicroseconds,
		),
	)
}
