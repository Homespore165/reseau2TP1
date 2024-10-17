package main

import (
	"database/sql"
	"fmt"
	"log"
)

var db *deckDB

type deckDB struct {
	db *sql.DB
}

var dbRequestChannel = make(chan DBRequest)

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

func DbManager() {
	for req := range dbRequestChannel {
		var response DBResponse
		switch req.QueryType {
		case "createDeck":
			err := _createDeck(req.Parameters[0].(string))
			response = DBResponse{Result: nil, Err: err}
		case "deckExists":
			exists := _deckExists(req.Parameters[0].(string))
			response = DBResponse{Result: exists, Err: nil}
		case "createCard":
			err := _createCard(req.Parameters[0].(string), req.Parameters[1].(int), req.Parameters[2].(int), req.Parameters[3].(int))
			response = DBResponse{Result: nil, Err: err}
		case "remainingCards":
			remaining := _remainingCards(req.Parameters[0].(string))
			response = DBResponse{Result: remaining, Err: nil}
		case "addCard":
			err := _addCard(req.Parameters[0].(string), req.Parameters[1].(Card))
			response = DBResponse{Result: nil, Err: err}
		case "highestPosition":
			position := _highestPosition(req.Parameters[0].(string))
			response = DBResponse{Result: position, Err: nil}
		case "drawCards":
			cards, err := _drawCards(req.Parameters[0].(string), req.Parameters[1].(int))
			response = DBResponse{Result: cards, Err: err}
		case "shuffleDeck":
			err := _shuffleDeck(req.Parameters[0].(string))
			response = DBResponse{Result: nil, Err: err}
		case "positionsCollide":
			collide := _positionsCollide(req.Parameters[0].(string))
			response = DBResponse{Result: collide, Err: nil}
		case "drawnCards":
			cards, err := _drawnCards(req.Parameters[0].(string), req.Parameters[1].(int))
			response = DBResponse{Result: cards, Err: err}
		case "comingCards":
			cards, err := _comingCards(req.Parameters[0].(string), req.Parameters[1].(int))
			response = DBResponse{Result: cards, Err: err}

		default:
			response = DBResponse{Err: fmt.Errorf("unknown query type")}
		}
		req.Response <- response
	}
}

func _createDeck(id string) error {
	_, err := db.db.Exec("INSERT INTO decks (id) VALUES (?)", id)
	return err
}

func createDeck(id string) error {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "createDeck",
		Parameters: []interface{}{id},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Err
}

func _deckExists(id string) bool {
	var exists bool
	err := db.db.QueryRow("SELECT EXISTS(SELECT 1 FROM decks WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func deckExists(id string) bool {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "deckExists",
		Parameters: []interface{}{id},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.(bool)
}

func _createCard(deckId string, suit int, rank int, position int) error {
	_, err := db.db.Exec("INSERT INTO cards (deck_id, rank, suit, position) VALUES (?, ?, ?, ?)", deckId, rank, suit, position)
	return err
}

func createCard(deckId string, suit int, rank int, position int) error {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "createCard",
		Parameters: []interface{}{deckId, suit, rank, position},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Err
}

func _remainingCards(deckId string) int {
	var remaining int
	err := db.db.QueryRow("SELECT COUNT(*) FROM cards WHERE deck_id = ? AND drawnDate IS NULL", deckId).Scan(&remaining)
	if err != nil {
		return 0
	}
	return remaining
}

func remainingCards(deckId string) int {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "remainingCards",
		Parameters: []interface{}{deckId},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.(int)
}

func _addCard(deckId string, card Card) error {
	pos := highestPosition(deckId) + 1
	_, err := db.db.Exec("INSERT INTO cards (deck_id, rank, suit, position) VALUES (?, ?, ?, ?)", deckId, card.Rank, card.Suit, pos)
	return err
}

func addCard(deckId string, card Card) error {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "addCard",
		Parameters: []interface{}{deckId, card},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Err
}

func _highestPosition(id string) int {
	var position int
	err := db.db.QueryRow("SELECT MAX(position) FROM cards WHERE deck_id = ?", id).Scan(&position)
	if err != nil {
		return 0
	}
	return position
}

func highestPosition(id string) int {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "highestPosition",
		Parameters: []interface{}{id},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.(int)
}

func _drawCards(deckId string, count int) ([]Card, error) {
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

func drawCards(deckId string, count int) ([]Card, error) {
	println("drawCards")
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "drawCards",
		Parameters: []interface{}{deckId, count},
		Response:   responseChannel,
	}
	println("drawCards")

	response := <-responseChannel
	return response.Result.([]Card), response.Err
}

func _shuffleDeck(deckId string) error {
	query := "UPDATE cards SET position = RANDOM() WHERE deck_id = ? AND drawnDate IS NULL"
	_, err := db.db.Exec(query, deckId)
	for positionsCollide(deckId) {
		_, err = db.db.Exec(query, deckId)
	}
	return err
}

func shuffleDeck(deckId string) error {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "shuffleDeck",
		Parameters: []interface{}{deckId},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Err
}

func _positionsCollide(deckId string) bool {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM cards WHERE deck_id = ? GROUP BY position HAVING COUNT(*) > 1", deckId).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func positionsCollide(deckId string) bool {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "positionsCollide",
		Parameters: []interface{}{deckId},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.(bool)
}

func _drawnCards(deckId string, nbrCards int) ([]Card, error) {
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

func drawnCards(deckId string, nbrCards int) ([]Card, error) {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "drawnCards",
		Parameters: []interface{}{deckId, nbrCards},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.([]Card), response.Err
}

func _comingCards(deckId string, nbrCards int) ([]Card, error) {
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

func comingCards(deckId string, nbrCards int) ([]Card, error) {
	responseChannel := make(chan DBResponse)
	dbRequestChannel <- DBRequest{
		QueryType:  "comingCards",
		Parameters: []interface{}{deckId, nbrCards},
		Response:   responseChannel,
	}

	response := <-responseChannel
	return response.Result.([]Card), response.Err
}
