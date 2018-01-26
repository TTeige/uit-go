package models

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func OpenDatabase(source string) (*sql.DB, error) {
	db, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
