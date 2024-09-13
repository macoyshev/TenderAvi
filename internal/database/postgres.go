package database

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (db *sql.DB, err error) {
	connStr, ok := os.LookupEnv("POSTGRES_CONN")
	if !ok {
		errMsg := "no POSTGRES_CONN environment variable"
		slog.Error(errMsg)
		err = errors.New(errMsg)
		return
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	err = db.Ping()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	return
}
