package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	sql *sql.DB
}

func DatabaseConnect(database, host, user, password string) (db *Database, err error) {

	sql, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database))
	if err != nil {

		return
	}
	db.sql = sql

	return
}

func (db *Database) NewClient(nick, lastseen string) (r sql.Result, err error) {

	err = db.sql.Ping()
	if err != nil {

		return
	}

	return db.sql.Exec("INSERT INTO clients VALUES (?)", nick, lastseen)
}

func (db *Database) UpdateClient(nick, lastseen string) (r sql.Result, err error) {

	err = db.sql.Ping()
	if err != nil {

		return
	}

	return db.sql.Exec("UPDATE clients SET (?) WHERE nick=?", lastseen, nick)
}

func (db *Database) GetClient(nick string) (lastseen string, err error) {

	err = db.sql.Ping()
	if err != nil {

		return
	}

	err = db.sql.QueryRow("SELECT lastseen FROM clients WHERE nick=?", nick).Scan(&lastseen)
	return
}
