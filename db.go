package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	user     = "postgres"
	password = "pass"
	host     = "prix.plus"
	dbname   = "admin"
	sslmode  = "disable"
)

func InitDB() (*sql.DB, error) {

	dbinfo := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s",
		user, password, host, dbname, sslmode)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}

	// Testing DB connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
