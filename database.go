package main

import "database/sql"

var db *deckDB

type deckDB struct {
	db *sql.DB
}

func initDB() (*deckDB, error) {
	creationQuery := `CREATE TABLE IF NOT EXISTS decks (
		id TEXT PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		deck_id TEXT,
		rank INTEGER,
		suit ENUM('h', 'd', 'c', 's')
);`

	db, err := sql.Open("sqlite3", "./deck.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(creationQuery)
	if err != nil {
		return nil, err
	}

	return &deckDB{db}, nil
}
