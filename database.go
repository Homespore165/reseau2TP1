package main

import (
	"database/sql"
	"log"
)

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
		suit INTEGER,
		position INTEGER,
		drawnDate TEXT DEFAULT NULL,
		FOREIGN KEY(deck_id) REFERENCES decks(id)
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

func createDeck(id string) error {
	_, err := db.db.Exec("INSERT INTO decks (id) VALUES (?)", id)
	return err
}

func deckExists(id string) bool {
	var exists bool
	err := db.db.QueryRow("SELECT EXISTS(SELECT 1 FROM decks WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func createCard(deckId string, suit int, rank int, position int) error {
	_, err := db.db.Exec("INSERT INTO cards (deck_id, rank, suit, position) VALUES (?, ?, ?, ?)", deckId, rank, suit, position)
	return err
}

func remainingCards(deckId string) int {
	var remaining int
	err := db.db.QueryRow("SELECT COUNT(*) FROM cards WHERE deck_id = ? AND drawnDate IS NULL", deckId).Scan(&remaining)
	if err != nil {
		return 0
	}
	return remaining
}

func addCard(deckId string, card Card) error {
	pos := highestPosition(deckId) + 1
	_, err := db.db.Exec("INSERT INTO cards (deck_id, rank, suit, position) VALUES (?, ?, ?, ?)", deckId, card.Rank, card.Suit, pos)
	return err
}

func highestPosition(id string) int {
	var position int
	err := db.db.QueryRow("SELECT MAX(position) FROM cards WHERE deck_id = ?", id).Scan(&position)
	if err != nil {
		return 0
	}
	return position
}

func drawCards(deckId string, count int) ([]Card, error) {
	var cards []Card
	rows, err := db.db.Query("SELECT rank, suit FROM cards WHERE deck_id = ? AND drawnDate IS NULL ORDER BY position DESC LIMIT ?", deckId, count)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	for rows.Next() {
		var card Card
		err = rows.Scan(&card.Rank, &card.Suit)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	_, err = db.db.Exec("UPDATE cards SET drawnDate = datetime('now') FROM (SELECT id FROM cards WHERE deck_id = ? AND drawnDate IS NULL ORDER BY position DESC LIMIT ?) AS sub WHERE cards.id = sub.id", deckId, count)
	if err != nil {
		log.Fatal(err)
	}
	return cards, nil
}

func shuffleDeck(deckId string) error {
	query := "UPDATE cards SET position = RANDOM() WHERE deck_id = ? AND drawnDate IS NULL"
	_, err := db.db.Exec(query, deckId)
	for positionsCollide(deckId) {
		_, err = db.db.Exec(query, deckId)
	}
	return err
}

func positionsCollide(deckId string) bool {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM cards WHERE deck_id = ? GROUP BY position HAVING COUNT(*) > 1", deckId).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func drawnCards(deckId string, nbrCards int) ([]Card, error) {
	var cards []Card
	query := "SELECT rank, suit FROM cards WHERE deck_id = ? AND drawnDate IS NOT NULL ORDER BY drawnDate, position DESC LIMIT ?"
	rows, err := db.db.Query(query, deckId, nbrCards)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	for rows.Next() {
		var card Card
		err = rows.Scan(&card.Rank, &card.Suit)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func comingCards(deckId string, nbrCards int) ([]Card, error) {
	var cards []Card
	query := "SELECT rank, suit FROM cards WHERE deck_id = ? AND drawnDate IS NULL ORDER BY position DESC LIMIT ?"
	rows, err := db.db.Query(query, deckId, nbrCards)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	for rows.Next() {
		var card Card
		err = rows.Scan(&card.Rank, &card.Suit)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}
