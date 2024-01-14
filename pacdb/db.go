package pacdb

import (
	"database/sql"
	_ "github.com/glebarez/go-sqlite"
)

var (
	database *sql.DB
)

func connect() error {
	if database != nil {
		return nil
	}

	if db, err := sql.Open("sqlite", "/var/lib/pacdb/pacman.sqlite"); err != nil {
		return err
	} else {
		database = db
	}

	return nil
}

func disconnect() error {
	if database == nil {
		return nil
	}

	if err := database.Close(); err != nil {
		return err
	}

	database = nil

	return nil
}
