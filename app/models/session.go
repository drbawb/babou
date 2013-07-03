package models

import (
	"database/sql"
	"fmt"
	"github.com/drbawb/babou/lib/db"
)

const (
	SESSIONS_TABLE string = "sessions"
)

type Session struct {
	sessionId  int
	userId     int
	loginIp    int
	lastSeenIp int
}

func (s *Session) WriteFor(user *User, ipAddr string) error {
	//user.lazyLoad()
	//ensure user is populated or able to be populated

	//check if session exists; if so: update with last_seen_at and last_seen_ip
	updateSession := `UPDATE "` + SESSIONS_TABLE + `" 
		SET login_ip = $2
	WHERE user_id = $1`

	insertSession := `INSERT INTO "` + SESSIONS_TABLE + `"(user_id, login_ip) 
	VALUES($1, $2)`

	// try parse IP

	dba := func(dbConn *sql.DB) error {
		result, err := dbConn.Exec(updateSession, user.UserId, "::1")
		if err != nil {
			fmt.Printf("error updating user's sessions: %s \n", err.Error())
			return err
		}

		rowsUpdated, err := result.RowsAffected()
		if err != nil {
			fmt.Printf("Error checking # rows affected by user session update: %s", err.Error())
			return err
		}
		// First update previous session record if it can be found.
		if rowsUpdated > 0 {
			fmt.Printf("session update be OK <3 \n")
			return nil
		}

		// Then insert a new one.
		fmt.Printf("session can no be update. hmm. maybe ok maybe i insert.")
		result, err = dbConn.Exec(insertSession, user.UserId, "::1")
		if err != nil {
			fmt.Printf("error inserting user's sessions: %s", err.Error())
			return err
		}

		return nil
	}

	err := db.ExecuteFn(dba)
	return err
}
