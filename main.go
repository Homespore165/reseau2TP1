package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func deckRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	if len(parts) < 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if parts[1] == "new" {
		deckNewHandler(w, r)
	} else if deckExists(parts[1]) {
		deckIdHandler(w, r)
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func deckIdHandler(w http.ResponseWriter, r *http.Request) {
	// add, draw, shuffle, show/0, show/1
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	if len(parts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	switch parts[2] {
	case "add":
		addHandler(w, r)
	case "draw":
		drawHandler(w, r)
	case "shuffle":
		shuffleHandler(w, r)
	case "show":
		showHandler(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	deckId := parts[1]

	if len(parts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	switch parts[3] {
	case "0":
		nbrCards := 1
		if len(parts) == 5 {
			var err error
			nbrCards, err = strconv.Atoi(parts[4])
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		}
		jsonOutputDrawn(w, deckId, nbrCards)
	case "1":
		nbrCards := 1
		if len(parts) == 5 {
			var err error
			nbrCards, err = strconv.Atoi(parts[4])
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		}
		jsonOutputComing(w, deckId, nbrCards)
	default:
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func jsonOutputComing(w http.ResponseWriter, deckId string, nbrCards int) {
	cards, err := comingCards(deckId, nbrCards)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	output := draw{DeckID: deckId, Cards: cards, Remaining: remainingCards(deckId)}
	outputJson, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(outputJson); err != nil {
		return
	}
}

func jsonOutputDrawn(w http.ResponseWriter, deckId string, nbrCards int) {
	cards, err := drawnCards(deckId, nbrCards)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	output := draw{DeckID: deckId, Cards: cards, Remaining: remainingCards(deckId)}
	outputJson, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(outputJson); err != nil {
		return
	}
}

func shuffleHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	deckId := parts[1]

	err := shuffleDeck(deckId)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	jsonOutputDeck(w, deckId)
}

func drawHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	deckId := parts[1]

	nbrCards := 1
	if len(parts) == 4 {
		var err error
		nbrCards, err = strconv.Atoi(parts[3])
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	cards, err := drawCards(deckId, nbrCards)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	remaining := remainingCards(deckId)
	errMsg := ""
	if remaining == 0 && len(cards) < nbrCards {
		errMsg = "Deck is empty"
	}

	output := draw{DeckID: deckId, Cards: cards, Remaining: remaining, Error: errMsg}

	outputJson, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(outputJson); err != nil {
		return
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	deckId := parts[1]

	cards := r.URL.Query().Get("cards")
	if cards == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cardList := strings.Split(cards, ",")
	for _, card := range cardList {
		card, err := parseCardCode(card)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		err = addCard(deckId, card)
	}
	jsonOutputDeck(w, deckId)
}

func parseCardCode(cardCode string) (Card, error) {
	suitStr := strings.ToLower(string(cardCode[len(cardCode)-1]))
	validSuits := map[string]bool{"h": true, "d": true, "c": true, "s": true}
	if !validSuits[suitStr] {
		return Card{}, fmt.Errorf("Invalid suit")
	}
	suitMap := map[string]int{"h": 1, "d": 2, "c": 3, "s": 4}
	suit := suitMap[suitStr]

	rankStr := cardCode[:len(cardCode)-1]
	rankMap := map[string]int{"j": 11, "q": 12, "k": 13, "x": 0}
	rank, ok := rankMap[rankStr]
	if !ok {
		var err error
		rank, err = strconv.Atoi(rankStr)
		if err != nil {
			return Card{}, fmt.Errorf("Invalid rank")
		}
	}

	code := rankStr + strconv.Itoa(suit)
	image := fmt.Sprintf("/static/%s.svg", cardCode)
	return Card{Rank: rank, Suit: suit, Code: code, Image: image}, nil
}

func deckNewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	parts = parts[1:]
	includeJokers := false
	nbrPack := 1

	if len(parts) == 2 {
		deckId := createNewDeck(nbrPack, includeJokers)
		jsonOutputDeck(w, deckId)
	} else if _, err := strconv.Atoi(parts[2]); err == nil {
		nbrPack, _ = strconv.Atoi(parts[2])
		if len(parts) == 4 {
			switch parts[3] {
			case "jokers":
				includeJokers = true
			default:
				http.Error(w, "Invalid request", http.StatusBadRequest)
			}
		}
		deckId := createNewDeck(nbrPack, includeJokers)
		jsonOutputDeck(w, deckId)
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func jsonOutputDeck(w http.ResponseWriter, deckId string) {
	remaining := remainingCards(deckId)
	deckJson, _ := json.Marshal(Deck{ID: deckId, Remaining: remaining})
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(deckJson); err != nil {
		return
	}
}

func createNewDeck(nbrPack int, includeJokers bool) string {
	deckId := uuid.New().String()
	for deckExists(deckId) {
		deckId = uuid.New().String()
	}
	err := createDeck(deckId)
	if err != nil {
		log.Fatal(err)
	}
	suits := []int{1, 2, 3, 4} // h, d, c, s
	ranks := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	pos := 0
	for i := 1; i < nbrPack+1; i++ {
		for _, suit := range suits {
			for _, rank := range ranks {
				err := createCard(deckId, suit, rank, pos)
				if err != nil {
					log.Fatal(err)
				}
				pos++
			}
		}
		if includeJokers {
			err := createCard(deckId, 1, 0, pos)
			pos++
			if err != nil {
				log.Fatal(err)
			}
			err = createCard(deckId, 4, 0, pos)
			pos++
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return deckId
}

func alternativeSolutionHandler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("https://www.420c56.drynish.synology.me/%s", r.URL.Path)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return
	}
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatal(err)
	}
	go DbManager()

	//http.HandleFunc("/", alternativeSolutionHandler)
	http.HandleFunc("/deck/", deckRouter)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
